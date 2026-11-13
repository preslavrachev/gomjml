package mjml

import (
	"context"
	"hash/maphash"
	"sync"
	"sync/atomic"
	"time"

	"github.com/preslavrachev/gomjml/mjml/debug"
)

// cachedAST wraps an MJML AST with a fixed expiration time.
// Entries are immutable once stored in the cache to avoid concurrent mutation.
type cachedAST struct {
	node    *MJMLNode
	expires time.Time
}

// ASTCache stores parsed MJML templates so that repeated renders of the same
// template skip parsing. Pass one to a render with WithASTCache.
//
// WithCache uses a package-wide ASTCache that is configured with
// SetASTCacheTTLOnce and SetASTCacheCleanupIntervalOnce. Create your own with
// NewASTCache when you want a cache with its own lifetime and settings, such
// as one per server or one per test.
//
// MEMORY MANAGEMENT STRATEGY:
//   - Fixed TTL expiration (no LRU) keeps implementation simple and predictable
//   - Background cleanup prevents unbounded memory growth
//   - No size limits - monitor memory usage in production environments
//   - Cache grows between cleanup cycles, then shrinks during cleanup
//
// CONCURRENCY ARCHITECTURE:
//   - sync.Map for the cache itself (optimized for high read/low write workloads)
//   - Singleflight pattern prevents duplicate parsing under high concurrency
//   - Multiple mutexes to minimize lock contention and prevent deadlocks
//
// When to Use Caching:
//   - High-volume applications rendering the same templates repeatedly
//   - Web servers with template reuse patterns
//   - Batch processing with repeated template rendering
//   - Applications where parsing time > rendering time
//
// When NOT to Use Caching:
//   - Single-use template rendering
//   - Memory-constrained environments
//   - Applications with constantly changing templates
//   - Short-lived processes where cache warmup overhead > benefits
//
// An ASTCache is safe for concurrent use. The zero value is an empty cache
// with the default settings.
type ASTCache struct {
	entries sync.Map // map[uint64]*cachedAST

	configMu        sync.RWMutex // protects ttl and cleanupInterval
	ttl             time.Duration
	cleanupInterval time.Duration

	cleanupMu     sync.Mutex // protects the cleanup goroutine lifecycle
	cleanupCancel context.CancelFunc
	cleanupDone   chan struct{}
	closed        atomic.Bool

	flights singleflightGroup
}

const defaultASTCacheTTL = 5 * time.Minute

// ASTCacheOption configures an ASTCache created by NewASTCache.
type ASTCacheOption func(*ASTCache)

// WithASTCacheTTL sets how long a parsed template stays cached. The default is
// 5 minutes. Non-positive values are ignored.
func WithASTCacheTTL(d time.Duration) ASTCacheOption {
	return func(c *ASTCache) {
		if d > 0 {
			c.ttl = d
		}
	}
}

// WithASTCacheCleanupInterval sets how often expired entries are removed. The
// default is half the TTL. Non-positive values are ignored.
func WithASTCacheCleanupInterval(d time.Duration) ASTCacheOption {
	return func(c *ASTCache) {
		if d > 0 {
			c.cleanupInterval = d
		}
	}
}

// NewASTCache returns an empty ASTCache. Its cleanup goroutine starts on first
// use and runs until Close. Close it when done; an unclosed cache keeps its
// cleanup goroutine and entries alive.
func NewASTCache(opts ...ASTCacheOption) *ASTCache {
	c := &ASTCache{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// effectiveTTL resolves an unset or non-positive TTL, as in a zero-value
// cache, to the default. The caller holds configMu.
func (c *ASTCache) effectiveTTL() time.Duration {
	if c.ttl <= 0 {
		return defaultASTCacheTTL
	}
	return c.ttl
}

// effectiveCleanupInterval defaults to half the TTL. The caller holds configMu.
func (c *ASTCache) effectiveCleanupInterval() time.Duration {
	if c.cleanupInterval <= 0 {
		// max guards a 1ns TTL: time.NewTicker panics on a zero interval.
		return max(c.effectiveTTL()/2, time.Nanosecond)
	}
	return c.cleanupInterval
}

// Parse returns the AST for mjmlContent, parsing and caching it on a miss.
// Concurrent misses for the same template share a single parse. The returned
// AST is shared with other renders and must not be modified.
//
// After Close, Parse parses every call and caches nothing.
func (c *ASTCache) Parse(mjmlContent string) (*MJMLNode, error) {
	if c.closed.Load() {
		return parseMJMLContent(mjmlContent)
	}

	c.startCleanup()
	hash := hashTemplate(mjmlContent)
	if node, ok := c.lookup(hash); ok {
		if debug.Enabled() {
			debug.DebugLog("mjml", "parse-cache-hit", "Using cached MJML AST")
		}
		return node, nil
	}

	return c.flights.do(hash, func() (*MJMLNode, error) {
		// A flight that finished between the lookup above and do has already
		// stored the AST; without this check the late caller parses again.
		if node, ok := c.lookup(hash); ok {
			return node, nil
		}

		node, err := parseMJMLContent(mjmlContent)
		if err != nil {
			return nil, err
		}

		// Read TTL with proper synchronization for cache storage
		c.configMu.RLock()
		ttl := c.effectiveTTL()
		c.configMu.RUnlock()

		entry := &cachedAST{node: node, expires: time.Now().Add(ttl)}
		c.entries.Store(hash, entry)
		if c.closed.Load() {
			// Close ran while this parse was in flight; don't outlive its Clear.
			c.entries.CompareAndDelete(hash, entry)
		}
		return node, nil
	})
}

// lookup returns the unexpired AST stored under hash, removing it if expired.
func (c *ASTCache) lookup(hash uint64) (*MJMLNode, bool) {
	cached, found := c.entries.Load(hash)
	if !found {
		return nil, false
	}
	entry := cached.(*cachedAST)
	if time.Now().Before(entry.expires) {
		return entry.node, true
	}
	// CompareAndDelete leaves a fresh entry stored since the Load in place.
	c.entries.CompareAndDelete(hash, cached)
	return nil, false
}

// Len reports how many templates c holds, including expired ones that the
// cleanup has not removed yet.
func (c *ASTCache) Len() int {
	n := 0
	c.entries.Range(func(_, _ any) bool {
		n++
		return true
	})
	return n
}

// Close stops c's cleanup goroutine, waits for it to exit and drops every
// entry. Renders may keep using c afterwards; they parse without caching.
// Close is idempotent.
func (c *ASTCache) Close() {
	c.closed.Store(true)
	c.stopCleanup()
	c.entries.Clear()
}

// startCleanup launches a background goroutine to periodically remove expired cache entries.
//
// HOW it works: Starts a single goroutine with a ticker that scans the entire
// cache at regular intervals (default: half of cache TTL). Uses context cancellation
// for graceful shutdown via stopCleanup.
//
// Thread safety: Uses cleanupMu to ensure only one cleanup goroutine runs.
// The goroutine reads configuration with configMu to avoid races with
// configuration changes during startup.
func (c *ASTCache) startCleanup() {
	c.cleanupMu.Lock()
	defer c.cleanupMu.Unlock()
	if c.cleanupCancel != nil || c.closed.Load() {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	c.cleanupCancel = cancel
	c.cleanupDone = done
	go func() {
		defer close(done)

		// Read cleanup interval with proper synchronization
		c.configMu.RLock()
		interval := c.effectiveCleanupInterval()
		c.configMu.RUnlock()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				c.entries.Range(func(key, value any) bool {
					entry := value.(*cachedAST)
					if now.After(entry.expires) {
						c.entries.CompareAndDelete(key, value)
					}
					return true
				})
			case <-ctx.Done():
				return
			}
		}
	}()
}

// stopCleanup cancels the cleanup goroutine, if one is running, and waits for
// it to exit. A later Parse starts a new one unless c is closed.
func (c *ASTCache) stopCleanup() {
	c.cleanupMu.Lock()
	cancel, done := c.cleanupCancel, c.cleanupDone
	c.cleanupCancel = nil
	c.cleanupMu.Unlock()

	if cancel != nil {
		cancel()
	}
	// done outlives cancel so a concurrent second stop also waits for the exit.
	if done != nil {
		<-done
	}
}

// defaultASTCache backs WithCache and the package-level cache functions.
//
// WHY a package-wide default: Template parsing is expensive and templates are
// often reused. Sharing one cache across the process maximizes efficiency when
// multiple parts of an application render the same templates (e.g., web
// servers, batch processors) without having to pass a cache around.
var defaultASTCache = NewASTCache()

var (
	astCacheTTLOnce     sync.Once // ensures TTL is set only once
	astCacheCleanupOnce sync.Once // ensures cleanup interval set only once
)

// SetASTCacheTTLOnce sets the time-to-live for cached AST entries.
// Only the first call has an effect; subsequent calls are ignored.
// The cleanup interval defaults to half of this value unless explicitly set.
//
// It configures the package-wide cache used by WithCache. Use NewASTCache
// with WithASTCacheTTL to configure a cache of your own.
func SetASTCacheTTLOnce(d time.Duration) {
	astCacheTTLOnce.Do(func() {
		c := defaultASTCache
		c.configMu.Lock()
		c.ttl = d
		astCacheCleanupOnce.Do(func() {
			c.cleanupInterval = d / 2
		})
		c.configMu.Unlock()
	})
}

// SetASTCacheCleanupIntervalOnce sets how often expired AST cache entries
// are removed. Only the first call has an effect. By default this is half of
// the AST cache TTL.
//
// It configures the package-wide cache used by WithCache. Use NewASTCache
// with WithASTCacheCleanupInterval to configure a cache of your own.
func SetASTCacheCleanupIntervalOnce(d time.Duration) {
	astCacheCleanupOnce.Do(func() {
		c := defaultASTCache
		c.configMu.Lock()
		c.cleanupInterval = d
		c.configMu.Unlock()
	})
}

// StopASTCacheCleanup stops the background cleanup goroutine of the
// package-wide cache used by WithCache, and waits for it to exit. The next
// cached render starts it again. Caches created with NewASTCache are stopped
// with their Close method.
func StopASTCacheCleanup() {
	defaultASTCache.stopCleanup()
}

type sfCall struct {
	wg  sync.WaitGroup
	res *MJMLNode
	err error
}

// singleflightGroup deduplicates concurrent parses of the same template
// within one ASTCache.
type singleflightGroup struct {
	mu    sync.Mutex         // protects calls
	calls map[uint64]*sfCall // tracks in-progress parse operations
}

// do executes fn while ensuring only one execution per hash at a time.
// Calls with the same hash wait for the first invocation to complete and receive its result.
//
// WHY this pattern: Template parsing is expensive (XML parsing + AST creation).
// Without singleflight, if 100 concurrent goroutines request the same template,
// all 100 would perform identical parsing work, wasting CPU and memory.
//
// HOW it works: The first caller to request a hash becomes the "worker" and executes fn.
// Subsequent callers with the same hash become "waiters" that block on a WaitGroup
// until the worker completes. All waiters then receive the worker's result (success or error).
//
// This prevents the "thundering herd" problem and ensures expensive operations
// are performed only once per unique input, regardless of concurrency level.
//
// Thread safety: Protected by mu for map operations. Each sfCall uses
// a WaitGroup to coordinate between the worker and waiters.
func (g *singleflightGroup) do(hash uint64, fn func() (*MJMLNode, error)) (*MJMLNode, error) {
	g.mu.Lock()
	if c, ok := g.calls[hash]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.res, c.err
	}
	if g.calls == nil {
		g.calls = make(map[uint64]*sfCall)
	}
	c := &sfCall{}
	c.wg.Add(1)
	g.calls[hash] = c
	g.mu.Unlock()

	defer func() {
		c.wg.Done()
		g.mu.Lock()
		delete(g.calls, hash)
		g.mu.Unlock()
	}()

	c.res, c.err = fn()
	return c.res, c.err
}

var (
	hashSeed             maphash.Seed // random seed for DoS protection
	templateHashSeedOnce sync.Once    // ensures seed is set only once
)

// hashTemplate returns a 64-bit hash of the MJML template using a package-wide seed.
// It avoids storing and comparing large strings when indexing cached entries.
//
// Security note: The package-wide seed prevents malicious inputs from causing
// hash collisions that could degrade cache performance to O(n) lookup times.
//
// Thread safety: Reading hashSeed is safe after templateHashSeedOnce.Do() completes.
// The seed is set once and never modified.
func hashTemplate(s string) uint64 {
	templateHashSeedOnce.Do(func() {
		hashSeed = maphash.MakeSeed()
	})
	var h maphash.Hash
	h.SetSeed(hashSeed)
	h.WriteString(s)
	return h.Sum64()
}

// parseAST parses mjmlContent, through the cache selected by opts when
// caching is enabled.
func parseAST(mjmlContent string, opts *RenderOpts) (*MJMLNode, error) {
	if !opts.UseCache {
		return parseMJMLContent(mjmlContent)
	}
	if opts.ASTCache != nil {
		return opts.ASTCache.Parse(mjmlContent)
	}
	return defaultASTCache.Parse(mjmlContent)
}

// parseMJMLContent calls ParseMJML with debug logging.
func parseMJMLContent(mjmlContent string) (*MJMLNode, error) {
	if debug.Enabled() {
		debug.DebugLog("mjml", "parse-start", "Starting MJML parsing")
	}
	node, err := ParseMJML(mjmlContent)
	if err != nil {
		if debug.Enabled() {
			debug.DebugLogError("mjml", "parse-error", "Failed to parse MJML", err)
		}
		return nil, err
	}
	if debug.Enabled() {
		debug.DebugLog("mjml", "parse-complete", "MJML parsing completed successfully")
	}
	return node, nil
}
