package mjml

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const scopedCacheTpl = `<mjml><mj-body><mj-section><mj-column><mj-text>hi</mj-text></mj-column></mj-section></mj-body></mjml>`

func newTestASTCache(t *testing.T, opts ...ASTCacheOption) *ASTCache {
	t.Helper()
	c := NewASTCache(opts...)
	t.Cleanup(c.Close)
	return c
}

func countParses(t *testing.T) *atomic.Int32 {
	t.Helper()
	var calls atomic.Int32
	orig := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		calls.Add(1)
		return orig(s)
	}
	t.Cleanup(func() { ParseMJML = orig })
	return &calls
}

// pinDefaultTTL stops an earlier test's short default TTL from expiring
// entries mid-test.
func pinDefaultTTL(t *testing.T) {
	t.Helper()
	c := defaultASTCache
	c.configMu.Lock()
	orig := c.ttl
	c.ttl = time.Hour
	c.configMu.Unlock()
	t.Cleanup(func() {
		c.configMu.Lock()
		c.ttl = orig
		c.configMu.Unlock()
	})
}

func cleanupDone(c *ASTCache) chan struct{} {
	c.cleanupMu.Lock()
	defer c.cleanupMu.Unlock()
	return c.cleanupDone
}

func mustRenderAST(t *testing.T, tpl string, opts ...RenderOption) *MJMLNode {
	t.Helper()
	res, err := RenderWithAST(tpl, opts...)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	return res.AST
}

func TestNewASTCacheOptions(t *testing.T) {
	tests := []struct {
		name         string
		opts         []ASTCacheOption
		wantTTL      time.Duration
		wantInterval time.Duration
	}{
		{"defaults", nil, 5 * time.Minute, 5 * time.Minute / 2},
		{"interval follows ttl", []ASTCacheOption{WithASTCacheTTL(time.Minute)}, time.Minute, 30 * time.Second},
		{"explicit interval", []ASTCacheOption{WithASTCacheCleanupInterval(time.Second), WithASTCacheTTL(time.Minute)}, time.Minute, time.Second},
		{"non-positive ignored", []ASTCacheOption{WithASTCacheTTL(0), WithASTCacheCleanupInterval(-time.Second)}, 5 * time.Minute, 5 * time.Minute / 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestASTCache(t, tt.opts...)
			ttl, interval := c.effectiveTTL(), c.effectiveCleanupInterval()
			if ttl != tt.wantTTL || interval != tt.wantInterval {
				t.Fatalf("got ttl=%v interval=%v, want ttl=%v interval=%v",
					ttl, interval, tt.wantTTL, tt.wantInterval)
			}
		})
	}
}

func TestASTCacheInstancesDoNotShareEntries(t *testing.T) {
	resetASTCache()
	t.Cleanup(resetASTCache)
	calls := countParses(t)
	a, b := newTestASTCache(t), newTestASTCache(t)

	a1 := mustRenderAST(t, scopedCacheTpl, WithASTCache(a))
	b1 := mustRenderAST(t, scopedCacheTpl, WithASTCache(b))
	a2 := mustRenderAST(t, scopedCacheTpl, WithASTCache(a))
	b2 := mustRenderAST(t, scopedCacheTpl, WithASTCache(b))

	if got := calls.Load(); got != 2 {
		t.Fatalf("expected one parse per instance, got %d", got)
	}
	if a1 != a2 || b1 != b2 {
		t.Fatal("expected each instance to reuse its own AST")
	}
	if a1 == b1 {
		t.Fatal("expected instances to hold separate ASTs")
	}
	if a.Len() != 1 || b.Len() != 1 {
		t.Fatalf("expected 1 entry per instance, got a=%d b=%d", a.Len(), b.Len())
	}
	if n := defaultASTCache.Len(); n != 0 {
		t.Fatalf("expected default cache untouched, got %d entries", n)
	}
}

func TestASTCacheTTLIsPerInstance(t *testing.T) {
	defaultTTL := defaultASTCache.ttl
	short := newTestASTCache(t, WithASTCacheTTL(50*time.Millisecond))
	long := newTestASTCache(t, WithASTCacheTTL(time.Hour))

	s1, err := short.Parse(scopedCacheTpl)
	if err != nil {
		t.Fatalf("parse short: %v", err)
	}
	l1, err := long.Parse(scopedCacheTpl)
	if err != nil {
		t.Fatalf("parse long: %v", err)
	}
	time.Sleep(60 * time.Millisecond)
	s2, _ := short.Parse(scopedCacheTpl)
	l2, _ := long.Parse(scopedCacheTpl)

	if s1 == s2 {
		t.Fatal("expected short-TTL entry to expire")
	}
	if l1 != l2 {
		t.Fatal("expected long-TTL entry to survive")
	}
	if defaultASTCache.ttl != defaultTTL {
		t.Fatalf("instance TTL leaked into default cache: %v", defaultASTCache.ttl)
	}
}

func TestASTCacheCloseStopsOnlyItsCleanup(t *testing.T) {
	resetASTCache()
	t.Cleanup(resetASTCache)
	pinDefaultTTL(t)
	a, b := newTestASTCache(t), newTestASTCache(t)

	mustRenderAST(t, scopedCacheTpl, WithASTCache(a))
	bAST := mustRenderAST(t, scopedCacheTpl, WithASTCache(b))
	defaultAST := mustRenderAST(t, scopedCacheTpl, WithCache())

	aDone, bDone, defaultDone := cleanupDone(a), cleanupDone(b), cleanupDone(defaultASTCache)
	if aDone == nil || bDone == nil || defaultDone == nil {
		t.Fatal("expected every cache to have started its cleanup goroutine")
	}

	a.Close()

	select {
	case <-aDone:
	default:
		t.Fatal("Close returned before the cleanup goroutine exited")
	}
	for name, done := range map[string]chan struct{}{"other instance": bDone, "default": defaultDone} {
		select {
		case <-done:
			t.Fatalf("closing one instance stopped the %s cleanup", name)
		default:
		}
	}

	calls := countParses(t)
	mustRenderAST(t, scopedCacheTpl, WithASTCache(a))
	mustRenderAST(t, scopedCacheTpl, WithASTCache(a))
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected a closed cache to parse every render, got %d parses", got)
	}
	if a.Len() != 0 {
		t.Fatalf("expected a closed cache to hold nothing, got %d entries", a.Len())
	}
	if cleanupDone(a) != aDone {
		t.Fatal("expected a closed cache not to restart its cleanup")
	}

	if mustRenderAST(t, scopedCacheTpl, WithASTCache(b)) != bAST {
		t.Fatal("expected the other instance to keep its entry")
	}
	if mustRenderAST(t, scopedCacheTpl, WithCache()) != defaultAST {
		t.Fatal("expected the default cache to keep its entry")
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("expected cache hits after closing another instance, got %d parses", got)
	}

	a.Close() // idempotent
}

func TestASTCacheSingleflightIsPerInstance(t *testing.T) {
	entered := make(chan struct{}, 16)
	release := make(chan struct{})
	releaseOnce := sync.OnceFunc(func() { close(release) })
	defer releaseOnce()

	orig := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		entered <- struct{}{}
		<-release
		return orig(s)
	}
	defer func() { ParseMJML = orig }()

	a, b := newTestASTCache(t), newTestASTCache(t)
	caches := []*ASTCache{a, b}
	const perCache = 3
	results := make([][]*MJMLNode, len(caches))

	var wg sync.WaitGroup
	for i, c := range caches {
		results[i] = make([]*MJMLNode, perCache)
		for j := range perCache {
			wg.Add(1)
			go func() {
				defer wg.Done()
				node, err := c.Parse(scopedCacheTpl)
				if err != nil {
					t.Errorf("parse: %v", err)
				}
				results[i][j] = node
			}()
		}
	}

	for range caches {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			releaseOnce()
			wg.Wait()
			t.Fatal("expected each instance to parse on its own; one waited on the other's flight")
		}
	}
	releaseOnce()
	wg.Wait()

	select {
	case <-entered:
		t.Fatal("expected a single parse per instance")
	default:
	}
	for i, nodes := range results {
		for _, n := range nodes {
			if n == nil || n != nodes[0] {
				t.Fatalf("cache %d: expected all callers to share one AST", i)
			}
		}
		if caches[i].Len() != 1 {
			t.Fatalf("cache %d: expected 1 entry, got %d", i, caches[i].Len())
		}
	}
	if results[0][0] == results[1][0] {
		t.Fatal("expected instances to hold separate ASTs")
	}
}

func TestASTCacheConcurrentRenders(t *testing.T) {
	templates := []string{tplWithColor("red"), tplWithColor("blue"), scopedCacheTpl}
	for _, name := range []string{"austin-global-attributes", "austin-group-component", "austin-buttons"} {
		src, err := os.ReadFile(filepath.Join("testdata", name+".mjml"))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		templates = append(templates, string(src))
	}
	want := make([]string, len(templates))
	for i, tpl := range templates {
		html, err := Render(tpl)
		if err != nil {
			t.Fatalf("uncached render %d: %v", i, err)
		}
		want[i] = html
	}

	c := newTestASTCache(t)
	var wg sync.WaitGroup
	for g := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for k := range 4 * len(templates) {
				i := (g + k) % len(templates)
				got, err := Render(templates[i], WithASTCache(c))
				if err != nil {
					t.Errorf("render %d: %v", i, err)
					return
				}
				if got != want[i] {
					t.Errorf("template %d: cached render differs from uncached render", i)
					return
				}
			}
		}()
	}
	wg.Wait()

	if c.Len() != len(templates) {
		t.Fatalf("expected %d entries, got %d", len(templates), c.Len())
	}
}

func TestWithASTCacheNilUsesDefault(t *testing.T) {
	resetASTCache()
	t.Cleanup(resetASTCache)
	pinDefaultTTL(t)

	first := mustRenderAST(t, scopedCacheTpl, WithASTCache(nil))
	if mustRenderAST(t, scopedCacheTpl, WithCache()) != first {
		t.Fatal("expected WithASTCache(nil) to use the default cache")
	}
}

func TestZeroValueASTCache(t *testing.T) {
	var c ASTCache
	defer c.Close()
	first, err := c.Parse(scopedCacheTpl)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	time.Sleep(20 * time.Millisecond) // a zero cleanup interval panics the ticker goroutine
	if again, _ := c.Parse(scopedCacheTpl); again != first {
		t.Fatal("expected a zero-value cache to cache with the default TTL")
	}
}

func TestASTCacheCloseDuringParseKeepsNothing(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	orig := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		close(entered)
		<-release
		return orig(s)
	}
	defer func() { ParseMJML = orig }()

	c := NewASTCache()
	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := c.Parse(scopedCacheTpl); err != nil {
			t.Errorf("parse: %v", err)
		}
	}()
	<-entered
	c.Close()
	close(release)
	<-done

	if n := c.Len(); n != 0 {
		t.Fatalf("expected a parse finishing after Close to leave nothing, got %d entries", n)
	}
}
