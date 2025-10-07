package mjml

import (
	"context"
	"fmt"
	"hash/maphash"
	"io"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/constants"
	"github.com/preslavrachev/gomjml/mjml/debug"
	"github.com/preslavrachev/gomjml/mjml/fonts"
	"github.com/preslavrachev/gomjml/mjml/globals"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/mjml/styles"
	"github.com/preslavrachev/gomjml/parser"
)

// Type alias for convenience
type MJMLNode = parser.MJMLNode

// ParseMJML re-exports the parser function for convenience
var ParseMJML = parser.ParseMJML

// RenderOpts is an alias for convenience
type RenderOpts = options.RenderOpts

// RenderOption is a functional option for configuring MJML rendering
type RenderOption func(*RenderOpts)

// calculateOptimalBufferSize determines the optimal buffer size based on template complexity
func calculateOptimalBufferSize(mjmlContent string) int {
	mjmlSize := len(mjmlContent)
	componentCount := strings.Count(mjmlContent, "<mj-")

	// Prevent division by zero for empty MJML content
	if mjmlSize == 0 {
		// Return a reasonable default buffer size for empty input
		return 1024
	}

	// Calculate component density (components per 1000 characters)
	complexity := float64(componentCount) / float64(mjmlSize) * 1000

	if complexity > 10 {
		// Very dense template - needs more buffer per component
		return mjmlSize*5 + componentCount*180
	} else if complexity > 5 {
		// Medium density - balanced approach
		return mjmlSize*4 + componentCount*140
	} else {
		// Light template - more conservative
		return mjmlSize*3 + componentCount*100
	}
}

// cachedAST wraps an MJML AST with a fixed expiration time.
// Entries are immutable once stored in the cache to avoid concurrent mutation.
type cachedAST struct {
	node    *MJMLNode
	expires time.Time
}

// Global cache state and synchronization primitives.
//
// DESIGN PHILOSOPHY:
// The caching system uses global state to share parsed templates across all render
// operations within a process. This provides maximum cache efficiency but requires
// careful synchronization for thread safety.
//
// MEMORY MANAGEMENT STRATEGY:
// - Fixed TTL expiration (no LRU) keeps implementation simple and predictable
// - Background cleanup prevents unbounded memory growth
// - No size limits - monitor memory usage in production environments
// - Cache grows between cleanup cycles, then shrinks during cleanup
//
// CONCURRENCY ARCHITECTURE:
// - sync.Map for the cache itself (optimized for high read/low write workloads)
// - Singleflight pattern prevents duplicate parsing under high concurrency
// - Multiple mutexes to minimize lock contention and prevent deadlocks
//
// WHY global state: Template parsing is expensive and templates are often reused.
// Process-wide caching maximizes efficiency when multiple parts of an application
// render the same templates (e.g., web servers, batch processors).
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
var (
	// Cache storage and configuration
	astCache                sync.Map          // map[uint64]*cachedAST - main cache storage
	astCacheTTL             = 5 * time.Minute // default expiration time
	astCacheTTLOnce         sync.Once         // ensures TTL is set only once
	astCacheCleanupInterval = astCacheTTL / 2 // how often to run cleanup
	astCacheCleanupOnce     sync.Once         // ensures cleanup interval set only once

	// Cache lifecycle management
	cacheCleanupMutex sync.Mutex         // protects cleanup goroutine lifecycle
	cleanupCancel     context.CancelFunc // cancels background cleanup
	cacheConfigMutex  sync.RWMutex       // protects TTL/interval reads during startup

	// Template hashing for cache keys
	hashSeed             maphash.Seed // random seed for DoS protection
	templateHashSeedOnce sync.Once    // ensures seed is set only once

	// Singleflight deduplication
	sfMutex sync.Mutex                 // protects singleflight map operations
	sfCalls = make(map[uint64]*sfCall) // tracks in-progress parse operations
)

// SetASTCacheTTLOnce sets the time-to-live for cached AST entries.
// Only the first call has an effect; subsequent calls are ignored.
// The cleanup interval defaults to half of this value unless explicitly set.
func SetASTCacheTTLOnce(d time.Duration) {
	astCacheTTLOnce.Do(func() {
		cacheConfigMutex.Lock()
		astCacheTTL = d
		astCacheCleanupOnce.Do(func() {
			astCacheCleanupInterval = d / 2
		})
		cacheConfigMutex.Unlock()
	})
}

// SetASTCacheCleanupIntervalOnce sets how often expired AST cache entries
// are removed. Only the first call has an effect. By default this is half of
// the AST cache TTL.
func SetASTCacheCleanupIntervalOnce(d time.Duration) {
	astCacheCleanupOnce.Do(func() {
		cacheConfigMutex.Lock()
		astCacheCleanupInterval = d
		cacheConfigMutex.Unlock()
	})
}

type sfCall struct {
	wg  sync.WaitGroup
	res *MJMLNode
	err error
}

// singleflightDo executes fn while ensuring only one execution per hash at a time.
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
// Thread safety: Protected by sfMu mutex for map operations. Each sfCall uses
// a WaitGroup to coordinate between the worker and waiters.
func singleflightDo(hash uint64, fn func() (*MJMLNode, error)) (*MJMLNode, error) {
	sfMutex.Lock()
	if c, ok := sfCalls[hash]; ok {
		sfMutex.Unlock()
		c.wg.Wait()
		return c.res, c.err
	}
	c := &sfCall{}
	c.wg.Add(1)
	sfCalls[hash] = c
	sfMutex.Unlock()

	defer func() {
		c.wg.Done()
		sfMutex.Lock()
		delete(sfCalls, hash)
		sfMutex.Unlock()
	}()

	c.res, c.err = fn()
	return c.res, c.err
}

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

// parseAST handles MJML parsing with optional caching.
func parseAST(mjmlContent string, useCache bool) (*MJMLNode, error) {
	if !useCache {
		debug.DebugLog("mjml", "parse-start", "Starting MJML parsing")
		node, err := ParseMJML(mjmlContent)
		if err != nil {
			debug.DebugLogError("mjml", "parse-error", "Failed to parse MJML", err)
			return nil, err
		}
		debug.DebugLog("mjml", "parse-complete", "MJML parsing completed successfully")
		return node, nil
	}

	startASTCacheCleanup()
	hash := hashTemplate(mjmlContent)
	if cached, found := astCache.Load(hash); found {
		entry := cached.(*cachedAST)
		if time.Now().Before(entry.expires) {
			debug.DebugLog("mjml", "parse-cache-hit", "Using cached MJML AST")
			return entry.node, nil
		}
		astCache.Delete(hash)
	}

	node, err := singleflightDo(hash, func() (*MJMLNode, error) {
		debug.DebugLog("mjml", "parse-start", "Starting MJML parsing")
		node, err := ParseMJML(mjmlContent)
		if err != nil {
			debug.DebugLogError("mjml", "parse-error", "Failed to parse MJML", err)
			return nil, err
		}
		debug.DebugLog("mjml", "parse-complete", "MJML parsing completed successfully")

		// Read TTL with proper synchronization for cache storage
		cacheConfigMutex.RLock()
		ttl := astCacheTTL
		cacheConfigMutex.RUnlock()

		astCache.Store(hash, &cachedAST{node: node, expires: time.Now().Add(ttl)})
		return node, nil
	})
	if err != nil {
		return nil, err
	}
	return node, nil
}

// startASTCacheCleanup launches a background goroutine to periodically remove expired cache entries.
//
// HOW it works: Starts a single goroutine with a ticker that scans the entire
// cache at regular intervals (default: half of cache TTL). Uses context cancellation
// for graceful shutdown via StopASTCacheCleanup().
//
// Thread safety: Uses cacheCleanupMutex to ensure only one cleanup goroutine runs.
// The goroutine reads configuration with cacheConfigMutex to avoid races with
// configuration changes during startup.
func startASTCacheCleanup() {
	cacheCleanupMutex.Lock()
	defer cacheCleanupMutex.Unlock()
	if cleanupCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	cleanupCancel = cancel
	go func() {
		// Read cleanup interval with proper synchronization
		cacheConfigMutex.RLock()
		interval := astCacheCleanupInterval
		cacheConfigMutex.RUnlock()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				now := time.Now()
				astCache.Range(func(key, value interface{}) bool {
					entry := value.(*cachedAST)
					if now.After(entry.expires) {
						astCache.Delete(key)
					}
					return true
				})
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StopASTCacheCleanup stops the background cache cleanup goroutine.
func StopASTCacheCleanup() {
	cacheCleanupMutex.Lock()
	defer cacheCleanupMutex.Unlock()
	if cleanupCancel != nil {
		cleanupCancel()
		cleanupCancel = nil
	}
}

// WithDebugTags enables or disables debug tag inclusion in the rendered output
func WithDebugTags(enabled bool) RenderOption {
	return func(opts *RenderOpts) {
		opts.DebugTags = enabled
	}
}

// WithCache enables AST caching
func WithCache() RenderOption {
	return func(opts *RenderOpts) {
		opts.UseCache = true
	}
}

// RenderResult contains both the rendered HTML and the MJML AST
type RenderResult struct {
	HTML string
	AST  *MJMLNode
}

var groupColumnClassOrderRegexp = regexp.MustCompile(`class="mj-outlook-group-fix (mj-column-(?:per|px)-[^" ]+)([^"]*)"`)

// RenderWithAST provides the internal MJML to HTML conversion function that returns both HTML and AST
func RenderWithAST(mjmlContent string, opts ...RenderOption) (*RenderResult, error) {
	startTime := time.Now()
	debug.DebugLogWithData("mjml", "render-start", "Starting MJML rendering", map[string]interface{}{
		"content_length": len(mjmlContent),
		"has_debug":      len(opts) > 0,
	})

	// Apply render options
	renderOpts := &RenderOpts{
		FontTracker: options.NewFontTracker(),
	}
	for _, opt := range opts {
		opt(renderOpts)
	}

	var validationErr *Error
	existingReporter := renderOpts.InvalidAttributeReporter
	renderOpts.InvalidAttributeReporter = func(tagName, attrName string, line int) {
		errDetail := ErrInvalidAttribute(tagName, attrName, line)
		if validationErr == nil {
			validationErr = errDetail
		} else {
			validationErr.Append(errDetail)
		}
		if existingReporter != nil {
			existingReporter(tagName, attrName, line)
		}
	}

	// Parse MJML using the parser package (with optional cache)
	ast, err := parseAST(mjmlContent, renderOpts.UseCache)
	if err != nil {
		return nil, err
	}

	// Initialize global attributes
	globalAttrs := globals.NewGlobalAttributes()

	// Process global attributes from head if it exists
	if headNode := ast.FindFirstChild("mj-head"); headNode != nil {
		globalAttrs.ProcessAttributesFromHead(headNode)
	}

	// Set the global attributes instance
	globals.SetGlobalAttributes(globalAttrs)

	// Create component tree
	debug.DebugLog("mjml", "component-tree-start", "Creating component tree from AST")
	component, err := CreateComponent(ast, renderOpts)
	if err != nil {
		debug.DebugLogError("mjml", "component-tree-error", "Failed to create component tree", err)
		return nil, err
	}
	debug.DebugLog("mjml", "component-tree-complete", "Component tree created successfully")

	// Render to HTML with optimized pre-allocation based on template complexity
	bufferSize := calculateOptimalBufferSize(mjmlContent)
	debug.DebugLogWithData("mjml", "render-html-start", "Starting HTML rendering", map[string]interface{}{
		"buffer_size": bufferSize,
	})
	var html strings.Builder
	html.Grow(bufferSize) // Pre-allocate with complexity-aware sizing

	renderStart := time.Now()
	err = component.Render(&html)
	if err != nil {
		debug.DebugLogError("mjml", "render-html-error", "Failed to render HTML", err)
		return nil, err
	}
	renderDuration := time.Since(renderStart).Milliseconds()

	htmlOutput := html.String()
	totalDuration := time.Since(startTime).Milliseconds()

	debug.DebugLogWithData("mjml", "render-complete", "MJML rendering completed", map[string]interface{}{
		"output_length":    len(htmlOutput),
		"render_time_ms":   renderDuration,
		"total_time_ms":    totalDuration,
		"expansion_factor": float64(len(htmlOutput)) / float64(len(mjmlContent)),
	})

	if validationErr != nil {
		return &RenderResult{
			HTML: htmlOutput,
			AST:  ast,
		}, *validationErr
	}

	return &RenderResult{
		HTML: htmlOutput,
		AST:  ast,
	}, nil
}

// Render provides the main MJML to HTML conversion function
func Render(mjmlContent string, opts ...RenderOption) (string, error) {
	// Reset navbar ID counter for deterministic IDs within each render
	components.ResetNavbarIDCounter()
	// Reset carousel ID counter for deterministic IDs within each render
	components.ResetCarouselIDCounter()

	result, err := RenderWithAST(mjmlContent, opts...)
	if result == nil {
		return "", err
	}
	normalizedHTML := normalizeGroupColumnClassOrder(result.HTML)
	return normalizedHTML, err
}

// RenderFromAST renders HTML from a pre-parsed AST
func RenderFromAST(ast *MJMLNode, opts ...RenderOption) (string, error) {
	// Apply render options
	renderOpts := &RenderOpts{}
	for _, opt := range opts {
		opt(renderOpts)
	}

	var validationErr *Error
	existingReporter := renderOpts.InvalidAttributeReporter
	renderOpts.InvalidAttributeReporter = func(tagName, attrName string, line int) {
		errDetail := ErrInvalidAttribute(tagName, attrName, line)
		if validationErr == nil {
			validationErr = errDetail
		} else {
			validationErr.Append(errDetail)
		}
		if existingReporter != nil {
			existingReporter(tagName, attrName, line)
		}
	}

	component, err := CreateComponent(ast, renderOpts)
	if err != nil {
		return "", err
	}

	html, err := RenderComponentString(component)
	if err != nil {
		return "", err
	}
	if validationErr != nil {
		return html, *validationErr
	}
	return html, nil
}

// NewFromAST creates a component from a pre-parsed AST (alias for CreateComponent)
func NewFromAST(ast *MJMLNode, opts ...RenderOption) (Component, error) {
	// Apply render options
	renderOpts := &RenderOpts{
		FontTracker: options.NewFontTracker(),
	}
	for _, opt := range opts {
		opt(renderOpts)
	}

	return CreateComponent(ast, renderOpts)
}

// normalizeGroupColumnClassOrder rewrites the mj-group column class ordering to match
// the canonical MJML output where the responsive width class precedes the Outlook fix
// helper. The rendering pipeline historically emitted the inverse ordering for
// internal helper tests, so we keep that behaviour in RenderWithAST while
// normalizing the public Render output to avoid integration diffs.
func normalizeGroupColumnClassOrder(input string) string {
	if !strings.Contains(input, "mj-outlook-group-fix mj-column-") {
		return input
	}
	return groupColumnClassOrderRegexp.ReplaceAllString(input, `class="$1 mj-outlook-group-fix$2"`)
}

// MJMLComponent represents the root MJML component
type MJMLComponent struct {
	*components.BaseComponent
	Head             *components.MJHeadComponent
	Body             *components.MJBodyComponent
	mobileCSSAdded   bool                   // Track if mobile CSS has been added
	columnClasses    map[string]styles.Size // Track column classes used in the document
	columnClassOrder []string               // Preserve insertion order of column classes
	carouselCSS      strings.Builder        // Collect carousel CSS from components
}

// RequestMobileCSS allows components to request mobile CSS to be added
func (c *MJMLComponent) RequestMobileCSS() {
	c.mobileCSSAdded = true
}

// RegisterCarouselCSS allows carousel components to register their CSS
func (c *MJMLComponent) RegisterCarouselCSS(css string) {
	c.carouselCSS.WriteString(css)
}

// collectCarouselCSS recursively collects carousel CSS from all components
func (c *MJMLComponent) collectCarouselCSS() {
	if c.Body != nil {
		c.collectCarouselCSSFromComponent(c.Body)
	}
}

// collectCarouselCSSFromComponent recursively collects carousel CSS from a component and its children
func (c *MJMLComponent) collectCarouselCSSFromComponent(comp Component) {
	// Check if this is a carousel component
	if carouselComp, ok := comp.(*components.MJCarouselComponent); ok {
		css := carouselComp.GenerateCSS()
		if css != "" {
			c.carouselCSS.WriteString(css)
		}
	}

	// Recursively process children based on component type
	switch v := comp.(type) {
	case *components.MJBodyComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJSectionComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJColumnComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJWrapperComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJGroupComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJSocialComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJAccordionComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJNavbarComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	case *components.MJCarouselComponent:
		for _, child := range v.Children {
			c.collectCarouselCSSFromComponent(child)
		}
	}
}

// hasCustomGlobalFonts checks if global attributes specify custom fonts
func (c *MJMLComponent) hasCustomGlobalFonts() bool {
	// Check if global attributes have specified font-family
	globalFontFamily := globals.GetGlobalAttribute("mj-all", "font-family")
	if globalFontFamily != "" && globalFontFamily != fonts.DefaultFontStack {
		return true
	}

	// Check if any text components have global font-family defined
	textFontFamily := globals.GetGlobalAttribute("mj-text", "font-family")
	if textFontFamily != "" && textFontFamily != fonts.DefaultFontStack {
		return true
	}

	return false
}

// prepareBodySiblings recursively sets up sibling relationships without rendering HTML
func (c *MJMLComponent) prepareBodySiblings(comp Component) {
	// Check specific component types that need to set up their children's sibling relationships
	switch v := comp.(type) {
	case *components.MJBodyComponent:
		// Body components set up their section children
		siblings := len(v.Children)
		rawSiblings := 0
		for _, child := range v.Children {
			if child.GetTagName() == "mj-raw" {
				rawSiblings++
			}
		}
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			child.SetSiblings(siblings)
			child.SetRawSiblings(rawSiblings)
			c.prepareBodySiblings(child)
		}
	case *components.MJSectionComponent:
		// Section components set up their column children
		siblings := len(v.Children)
		rawSiblings := 0
		for _, child := range v.Children {
			if child.GetTagName() == "mj-raw" {
				rawSiblings++
			}
		}
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			child.SetSiblings(siblings)
			child.SetRawSiblings(rawSiblings)
			c.prepareBodySiblings(child)
		}
	case *components.MJColumnComponent:
		// Column components set up their content children
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			c.prepareBodySiblings(child)
		}
	case *components.MJWrapperComponent:
		// Wrapper components set up their children
		for _, child := range v.Children {
			child.SetContainerWidth(v.GetEffectiveWidth())
			c.prepareBodySiblings(child)
		}
	case *components.MJGroupComponent:
		// Group components set up their children and distribute width equally
		columnCount := 0
		for _, child := range v.Children {
			if _, ok := child.(*components.MJColumnComponent); ok {
				columnCount++
			}
		}

		if columnCount > 0 {
			percentagePerColumn := 100.0 / float64(columnCount)

			for _, child := range v.Children {
				child.SetContainerWidth(v.GetEffectiveWidth())

				// Set width attributes on columns like the group's Render() method does
				if columnComp, ok := child.(*components.MJColumnComponent); ok {
					if columnComp.GetAttribute("width") == nil {
						percentageWidth := fmt.Sprintf("%.15f%%", percentagePerColumn)
						percentageWidth = strings.TrimRight(percentageWidth, "0")
						percentageWidth = strings.TrimRight(percentageWidth, ".")
						if !strings.HasSuffix(percentageWidth, "%") {
							percentageWidth += "%"
						}
						columnComp.Attrs["width"] = percentageWidth
					}
				}

				c.prepareBodySiblings(child)
			}
		} else {
			for _, child := range v.Children {
				child.SetContainerWidth(v.GetEffectiveWidth())
				c.prepareBodySiblings(child)
			}
		}
	}
}

// collectColumnClasses recursively collects all column classes used in the document
func (c *MJMLComponent) collectColumnClasses() {
	c.columnClasses = make(map[string]styles.Size)
	c.columnClassOrder = c.columnClassOrder[:0]
	if c.Body != nil {
		c.collectColumnClassesFromComponent(c.Body)
	}
}

func (c *MJMLComponent) registerColumnClass(className string, size styles.Size) {
	if _, exists := c.columnClasses[className]; !exists {
		c.columnClassOrder = append(c.columnClassOrder, className)
	}
	c.columnClasses[className] = size
}

// collectColumnClassesFromComponent recursively collects column classes from a component
func (c *MJMLComponent) collectColumnClassesFromComponent(comp Component) {
	// Check if this is a column component
	if columnComp, ok := comp.(*components.MJColumnComponent); ok {
		className, size := columnComp.GetColumnClass()
		c.registerColumnClass(className, size)
	}

	// Check specific component types that have children
	switch v := comp.(type) {
	case *components.MJBodyComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJSectionComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJColumnComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJWrapperComponent:
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	case *components.MJGroupComponent:
		// Register group's CSS class based on its width attribute
		groupWidth := v.GetAttribute("width")
		if groupWidth != nil && strings.HasSuffix(*groupWidth, "px") {
			// Parse pixel width and register pixel-based class
			var widthPx int
			fmt.Sscanf(*groupWidth, "%dpx", &widthPx)
			className := fmt.Sprintf("mj-column-px-%d", widthPx)
			c.registerColumnClass(className, styles.NewPixelSize(float64(widthPx)))
		} else {
			// Default to percentage-based class
			c.registerColumnClass("mj-column-per-100", styles.NewPercentSize(100))
		}

		// Also recurse into children to collect column classes
		for _, child := range v.Children {
			c.collectColumnClassesFromComponent(child)
		}
	}
}

// generateResponsiveCSS generates responsive CSS for collected column classes
func (c *MJMLComponent) generateResponsiveCSS() string {
	var css strings.Builder

	// Standard responsive media query
	css.WriteString("<style type=\"text/css\">@media only screen and (min-width:480px) {\n")
	// Deterministic ordering to match MRML byte output
	for _, className := range c.columnClassOrder {
		size := c.columnClasses[className]
		// Include both percentage and pixel-based classes
		css.WriteString("        .")
		css.WriteString(className)
		css.WriteString(" { width:")
		css.WriteString(size.String())
		css.WriteString(" !important; max-width: ")
		css.WriteString(size.String())
		css.WriteString("; }\n")
	}
	css.WriteString("      }</style>")

	// Mozilla-specific responsive media query
	css.WriteString(`<style media="screen and (min-width:480px)">`)
	first := true
	for _, className := range c.columnClassOrder {
		if !first {
			css.WriteByte(' ')
		}
		first = false

		size := c.columnClasses[className]
		// Include both percentage and pixel-based classes
		css.WriteString(`.moz-text-html .`)
		css.WriteString(className)
		css.WriteString(` { width:`)
		css.WriteString(size.String())
		css.WriteString(` !important; max-width: `)
		css.WriteString(size.String())
		css.WriteString(`; }`)
	}
	css.WriteString(`</style>`)

	return css.String()
}

// extractHeadMetadata collects document-level metadata from mj-head children such as title
// and custom font declarations. The extracted title is stored on the render options so that
// body-level rendering can access it for accessibility attributes (aria-label).
func (c *MJMLComponent) extractHeadMetadata() (string, []string) {
	title := ""
	customFonts := make([]string, 0)

	if c.Head == nil {
		if c.RenderOpts != nil {
			c.RenderOpts.Title = ""
		}
		return title, customFonts
	}

	for _, child := range c.Head.Children {
		switch comp := child.(type) {
		case *components.MJTitleComponent:
			title = strings.TrimSpace(comp.Node.Text)
		case *components.MJFontComponent:
			getAttr := func(name string) string {
				if attr := comp.GetAttribute(name); attr != nil {
					return *attr
				}
				return comp.GetDefaultAttribute(name)
			}

			fontName := getAttr("name")
			fontHref := getAttr("href")
			if fontName != "" && fontHref != "" {
				customFonts = append(customFonts, fontHref)
			}
		}
	}

	if c.RenderOpts != nil {
		c.RenderOpts.Title = title
	}

	return title, customFonts
}

// generateCustomStyles generates the final mj-style content tag (MRML lines 240-244)
func (c *MJMLComponent) generateCustomStyles() string {
	var content strings.Builder

	// Collect all mj-style content (MRML mj_style_iter equivalent)
	if c.Head != nil {
		for _, child := range c.Head.Children {
			if styleComp, ok := child.(*components.MJStyleComponent); ok {
				inlineAttr := ""
				if attr := styleComp.GetAttribute("inline"); attr != nil {
					inlineAttr = strings.ToLower(strings.TrimSpace(*attr))
				}
				if inlineAttr == "inline" {
					continue
				}

				text := strings.TrimSpace(styleComp.Node.Text)
				if text != "" {
					content.WriteString(text)
				}
			}
		}
	}

	// Only generate the style tag if there's content (MJML JS behavior)
	if content.Len() > 0 {
		return fmt.Sprintf(`<style type="text/css">%s</style>`, content.String())
	}
	return ""
}

// generateAccordionCSS generates the CSS styles needed for accordion functionality
func (c *MJMLComponent) generateAccordionCSS() string {
	return `<style type="text/css">noinput.mj-accordion-checkbox { display: block! important; }
@media yahoo, only screen and (min-width:0) {
  .mj-accordion-element { display:block; }
  input.mj-accordion-checkbox, .mj-accordion-less { display: none !important; }
  input.mj-accordion-checkbox+* .mj-accordion-title { cursor: pointer; touch-action: manipulation; -webkit-user-select: none; -moz-user-select: none; user-select: none; }
  input.mj-accordion-checkbox+* .mj-accordion-content { overflow: hidden; display: none; }
  input.mj-accordion-checkbox+* .mj-accordion-more { display: block !important; }
  input.mj-accordion-checkbox:checked+* .mj-accordion-content { display: block; }
  input.mj-accordion-checkbox:checked+* .mj-accordion-more { display: none !important; }
  input.mj-accordion-checkbox:checked+* .mj-accordion-less { display: block !important; }
}
.moz-text-html input.mj-accordion-checkbox+* .mj-accordion-title { cursor: auto; touch-action: auto; -webkit-user-select: auto; -moz-user-select: auto; user-select: auto; }
.moz-text-html input.mj-accordion-checkbox+* .mj-accordion-content { overflow: hidden; display: block; }
.moz-text-html input.mj-accordion-checkbox+* .mj-accordion-ico { display: none; }
@goodbye { @gmail }
</style>`
}

// generateNavbarCSS generates the CSS styles needed for navbar hamburger menu functionality
func (c *MJMLComponent) generateNavbarCSS() string {
	return `<style type="text/css">
        noinput.mj-menu-checkbox { display:block!important; max-height:none!important; visibility:visible!important; }
        @media only screen and (max-width:479px) {
          .mj-menu-checkbox[type="checkbox"] ~ .mj-inline-links { display:none!important; }
          .mj-menu-checkbox[type="checkbox"]:checked ~ .mj-inline-links,
          .mj-menu-checkbox[type="checkbox"] ~ .mj-menu-trigger { display:block!important; max-width:none!important; max-height:none!important; font-size:inherit!important; }
          .mj-menu-checkbox[type="checkbox"] ~ .mj-inline-links > a { display:block!important; }
          .mj-menu-checkbox[type="checkbox"]:checked ~ .mj-menu-trigger .mj-menu-icon-close { display:block!important; }
          .mj-menu-checkbox[type="checkbox"]:checked ~ .mj-menu-trigger .mj-menu-icon-open { display:none!important; }
        }
        </style>`
}

// generateCarouselCSS generates the CSS styles needed for carousel functionality
func (c *MJMLComponent) generateCarouselCSS() string {
	if c.carouselCSS.Len() == 0 {
		return ""
	}
	return "<style type=\"text/css\">" + c.carouselCSS.String() + "</style>"
}

// hasMobileCSSComponents recursively checks if any component needs mobile CSS
func (c *MJMLComponent) hasMobileCSSComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkComponentForMobileCSS(c.Body)
}

// hasTextComponents checks if the document contains any text-based components that need fonts
func (c *MJMLComponent) hasTextComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.hasTextComponentsRecursive(c.Body)
}

// hasSocialComponents checks if the MJML contains any social components
func (c *MJMLComponent) hasSocialComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		switch comp.GetTagName() {
		case "mj-social", "mj-social-element":
			return true
		}
		return false
	})
}

// hasButtonComponents checks if the MJML contains any button components
func (c *MJMLComponent) hasButtonComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		return comp.GetTagName() == "mj-button"
	})
}

// hasAccordionComponents checks if the MJML contains any accordion components
func (c *MJMLComponent) hasAccordionComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		switch comp.GetTagName() {
		case "mj-accordion", "mj-accordion-element", "mj-accordion-title", "mj-accordion-text":
			return true
		}
		return false
	})
}

// hasNavbarComponents checks if the MJML contains any navbar components
func (c *MJMLComponent) hasNavbarComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		switch comp.GetTagName() {
		case "mj-navbar", "mj-navbar-link":
			return true
		}
		return false
	})
}

// hasCarouselComponents checks if the MJML contains any carousel components
func (c *MJMLComponent) hasCarouselComponents() bool {
	if c.Body == nil {
		return false
	}
	return c.checkChildrenForCondition(c.Body, func(comp Component) bool {
		switch comp.GetTagName() {
		case "mj-carousel", "mj-carousel-image":
			return true
		}
		return false
	})
}

// shouldImportDefaultFonts determines if default fonts should be auto-imported
// based on detected fonts, social components presence, and custom global fonts
func (c *MJMLComponent) shouldImportDefaultFonts(detectedFonts []string, trackedFontsCount int, hasText, hasSocial, hasButtons bool, hasOnlyDefaultFonts bool) bool {
	if c.hasCustomGlobalFonts() {
		return false
	}

	// Social-only layouts don't trigger default font imports in MJML's reference output.
	if hasSocial && !hasText && !hasButtons {
		return false
	}

	if hasText || hasButtons {
		// If any fonts were tracked (including system fonts like Arial), don't import defaults
		// System fonts don't generate URLs but still count as explicit font usage
		if trackedFontsCount > 0 && len(detectedFonts) == 0 {
			return false
		}
		return len(detectedFonts) == 0 || hasOnlyDefaultFonts
	}

	return false
}

// hasTextComponentsRecursive recursively checks for text components
func (c *MJMLComponent) hasTextComponentsRecursive(component Component) bool {
	// Check if this component is a text component
	switch component.(type) {
	case *components.MJTextComponent, *components.MJButtonComponent:
		return true
	}

	// Check specific component types that have children
	return c.checkChildrenForCondition(component, c.hasTextComponentsRecursive)
}

// checkComponentForMobileCSS recursively checks a component and its children
func (c *MJMLComponent) checkComponentForMobileCSS(comp Component) bool {
	// Check if this component needs mobile CSS (currently only mj-image)
	if comp.GetTagName() == "mj-image" {
		return true
	}

	// Check specific component types that have children
	return c.checkChildrenForCondition(comp, c.checkComponentForMobileCSS)
}

// checkChildrenForCondition is a helper function that checks if any children of a component meet a condition
func (c *MJMLComponent) checkChildrenForCondition(component Component, condition func(Component) bool) bool {
	// Check all children recursively
	switch v := component.(type) {
	case *components.MJBodyComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJSectionComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJColumnComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJWrapperComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJGroupComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJSocialComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJAccordionComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	case *components.MJNavbarComponent:
		for _, child := range v.Children {
			if condition(child) || c.checkChildrenForCondition(child, condition) {
				return true
			}
		}
	}
	return false
}

func (c *MJMLComponent) GetTagName() string {
	return "mjml"
}

// Render implements optimized Writer-based rendering for MJMLComponent
func (c *MJMLComponent) Render(w io.StringWriter) error {
	debug.DebugLog("mjml-root", "render-start", "Starting root MJML component rendering")

	// First, prepare the body to establish sibling relationships without full rendering
	debug.DebugLog("mjml-root", "prepare-siblings", "Preparing body sibling relationships")
	if c.Body != nil {
		c.prepareBodySiblings(c.Body)
	}

	// Now collect column classes after sibling relationships are established
	debug.DebugLog("mjml-root", "collect-column-classes", "Collecting column classes for responsive CSS")
	c.collectColumnClasses()
	debug.DebugLogWithData("mjml-root", "column-classes-collected", "Column classes collected", map[string]interface{}{
		"class_count": len(c.columnClasses),
	})

	// Collect carousel CSS from all carousel components
	debug.DebugLog("mjml-root", "collect-carousel-css", "Collecting carousel CSS")
	c.collectCarouselCSS()

	// Extract head metadata (title, custom fonts) before rendering body so accessibility
	// attributes can access the document title during body rendering.
	title, customFonts := c.extractHeadMetadata()

	// Generate body content once for both font detection and final output
	debug.DebugLog("mjml-root", "render-body", "Rendering body content for font analysis and output")
	var bodyBuffer strings.Builder
	if c.Body != nil {
		if err := c.Body.Render(&bodyBuffer); err != nil {
			debug.DebugLogError("mjml-root", "render-body-error", "Failed to render body", err)
			return err
		}
	}
	bodyContent := bodyBuffer.String()
	debug.DebugLogWithData("mjml-root", "render-complete", "Body rendering completed", map[string]interface{}{
		"body_length": len(bodyContent),
	})

	// DOCTYPE and HTML opening - include attributes from MJML root element
	var langValue string
	if langAttr := c.GetAttribute("lang"); langAttr != nil {
		langValue = *langAttr
	} else {
		langValue = constants.LangUndetermined
	}

	var dirValue string
	if dirAttr := c.GetAttribute("dir"); dirAttr != nil {
		dirValue = *dirAttr
	} else {
		dirValue = constants.DirAuto
	}

	if _, err := w.WriteString(`<!doctype html><html lang="` + langValue + `" dir="` + dirValue + `" xmlns="http://www.w3.org/1999/xhtml" xmlns:v="urn:schemas-microsoft-com:vml" xmlns:o="urn:schemas-microsoft-com:office:office">`); err != nil {
		return err
	}
	if _, err := w.WriteString(`<head>`); err != nil {
		return err
	}

	if _, err := w.WriteString(`<title>` + title + `</title>`); err != nil {
		return err
	}
	if _, err := w.WriteString(`<!--[if !mso]><!--><meta http-equiv="X-UA-Compatible" content="IE=edge"><!--<![endif]-->`); err != nil {
		return err
	}
	if _, err := w.WriteString(`<meta http-equiv="Content-Type" content="text/html; charset=UTF-8">`); err != nil {
		return err
	}
	if _, err := w.WriteString(`<meta name="viewport" content="width=device-width,initial-scale=1">`); err != nil {
		return err
	}

	// Base CSS
	baseCSSText := `<style type="text/css">#outlook a { padding:0; }
      body { margin:0;padding:0;-webkit-text-size-adjust:100%;-ms-text-size-adjust:100%; }
      table, td { border-collapse:collapse;mso-table-lspace:0pt;mso-table-rspace:0pt; }
      img { border:0;height:auto;line-height:100%; outline:none;text-decoration:none;-ms-interpolation-mode:bicubic; }
      p { display:block;margin:13px 0; }</style>`
	if _, err := w.WriteString(baseCSSText); err != nil {
		return err
	}

	// MSO conditionals
	msoText := `<!--[if mso]>
    <noscript>
    <xml>
    <o:OfficeDocumentSettings>
      <o:AllowPNG/>
      <o:PixelsPerInch>96</o:PixelsPerInch>
    </o:OfficeDocumentSettings>
    </xml>
    </noscript>
    <![endif]--><!--[if lte mso 11]>
    <style type="text/css">
      .mj-outlook-group-fix { width:100% !important; }
    </style>
    <![endif]-->`
	if _, err := w.WriteString(msoText); err != nil {
		return err
	}

	// Font imports - auto-detect fonts from content and add custom fonts from mj-font
	var allFontsToImport []string

	// Add explicit custom fonts from mj-font components
	allFontsToImport = append(allFontsToImport, customFonts...)

	// Get fonts tracked during component rendering
	trackedFonts := c.RenderOpts.FontTracker.GetFonts()
	detectedFonts := fonts.ConvertFontFamiliesToURLs(trackedFonts)
	debug.DebugLogWithData(
		"font-detection",
		"component-tracking",
		"Fonts tracked from components",
		map[string]interface{}{
			"tracked_count": len(trackedFonts),
			"url_count":     len(detectedFonts),
			"fonts":         strings.Join(trackedFonts, ","),
		},
	)

	for _, detectedFont := range detectedFonts {
		// Only add if not already in custom fonts from mj-font
		alreadyExists := false
		for _, customFont := range customFonts {
			if customFont == detectedFont {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			allFontsToImport = append(allFontsToImport, detectedFont)
		}
	}

	// Also check for default fonts based on component presence (like MRML does)
	// Note: MRML only imports fonts when specific conditions are met, not just any text presence
	hasSocial := c.hasSocialComponents()
	hasButtons := c.hasButtonComponents()
	hasText := c.hasTextComponents()
	// Only auto-import default fonts if no fonts were already detected from content
	// This matches MRML's behavior: explicit fonts override default font imports
	// Also respect custom global fonts from mj-all attributes
	// Special case: social components with only default fonts should trigger Ubuntu fallback
	hasOnlyDefaultFonts := len(detectedFonts) == 1 && detectedFonts[0] == fonts.GetGoogleFontURL(fonts.DefaultFontStack)
	// Pass trackedFonts count to check if ANY fonts (including system fonts) were used
	if c.shouldImportDefaultFonts(detectedFonts, len(trackedFonts), hasText, hasSocial, hasButtons, hasOnlyDefaultFonts) {
		debug.DebugLogWithData(
			"font-detection",
			"check-defaults",
			"No content fonts detected, checking defaults",
			map[string]interface{}{
				"has_social": hasSocial,
			},
		)
		defaultFonts := fonts.DetectDefaultFonts(hasText, hasSocial, hasButtons)
		debug.DebugLogWithData("font-detection", "default-fonts", "Default fonts to import", map[string]interface{}{
			"count": len(defaultFonts),
			"fonts": strings.Join(defaultFonts, ","),
		})
		for _, defaultFont := range defaultFonts {
			// Only add if not already in existing fonts
			alreadyExists := false
			for _, existingFont := range allFontsToImport {
				if existingFont == defaultFont {
					alreadyExists = true
					break
				}
			}
			if !alreadyExists {
				allFontsToImport = append(allFontsToImport, defaultFont)
			}
		}
	} else {
		debug.DebugLogWithData("font-detection", "skip-defaults", "Skipping default fonts", map[string]interface{}{
			"detected_count": len(detectedFonts),
			"has_social":     hasSocial,
		})
	}

	// Generate font import HTML
	debug.DebugLogWithData("font-detection", "final-list", "Final fonts to import", map[string]interface{}{
		"total_count": len(allFontsToImport),
		"fonts":       strings.Join(allFontsToImport, ","),
	})
	if len(allFontsToImport) > 0 {
		fontImportsHTML := fonts.BuildFontsTags(allFontsToImport)
		if _, err := w.WriteString(fontImportsHTML); err != nil {
			return err
		}
	}

	// Dynamic responsive CSS based on collected column classes - only if we have columns
	if len(c.columnClasses) > 0 {
		responsiveCSS := c.generateResponsiveCSS()
		if _, err := w.WriteString(responsiveCSS); err != nil {
			return err
		}
	}

	// Mobile CSS - add only if components need it (following MRML pattern)
	if c.hasMobileCSSComponents() {
		mobileCSSText := `<style type="text/css">@media only screen and (max-width:479px) {
                table.mj-full-width-mobile { width: 100% !important; }
                td.mj-full-width-mobile { width: auto !important; }
            }
            </style>`
		if _, err := w.WriteString(mobileCSSText); err != nil {
			return err
		}
	}

	// Accordion CSS - add only if components need it (following MRML pattern)
	if c.hasAccordionComponents() {
		accordionCSSText := c.generateAccordionCSS()
		if _, err := w.WriteString(accordionCSSText); err != nil {
			return err
		}
	}

	// Navbar CSS - add only if components need it (following MRML pattern)
	if c.hasNavbarComponents() {
		navbarCSSText := c.generateNavbarCSS()
		if _, err := w.WriteString(navbarCSSText); err != nil {
			return err
		}
	}

	// Carousel CSS - add only if components need it (following MRML pattern)
	if c.hasCarouselComponents() {
		carouselCSSText := c.generateCarouselCSS()
		if _, err := w.WriteString(carouselCSSText); err != nil {
			return err
		}
	}

	// Custom styles from mj-style components (MRML lines 240-244)
	customStyles := c.generateCustomStyles()
	if _, err := w.WriteString(customStyles); err != nil {
		return err
	}
	if c.RenderOpts != nil && c.RenderOpts.RequireEmptyStyleTag && customStyles == "" {
		if _, err := w.WriteString(`<style type="text/css"></style>`); err != nil {
			return err
		}
		// Ensure we only emit the placeholder once per render.
		c.RenderOpts.RequireEmptyStyleTag = false
	}

	// Render mj-raw components inside head
	if c.Head != nil {
		for _, child := range c.Head.Children {
			if rawComp, ok := child.(*components.MJRawComponent); ok {
				if err := rawComp.Render(w); err != nil {
					return err
				}
			}
		}
	}

	if _, err := w.WriteString(`</head>`); err != nil {
		return err
	}

	// Body with background-color support (matching MRML's get_body_tag)
	var bodyStyles []string

	// Only add word-spacing:normal if there's actual body content to match MRML behavior
	if len(bodyContent) > 0 {
		bodyStyles = append(bodyStyles, "word-spacing:normal")
	}

	if c.Body != nil {
		if bgColor := c.Body.GetAttribute("background-color"); bgColor != nil && *bgColor != "" {
			bodyStyles = append(bodyStyles, "background-color:"+*bgColor)
		}
	}

	bodyTag := `<body>`
	if len(bodyStyles) > 0 {
		var styleBuilder strings.Builder
		for _, style := range bodyStyles {
			styleBuilder.WriteString(style)
			styleBuilder.WriteString(";")
		}
		bodyTag = `<body style="` + styleBuilder.String() + `">`
	}
	if _, err := w.WriteString(bodyTag); err != nil {
		return err
	}

	// Add preview text from head components right after body tag
	if c.Head != nil {
		for _, child := range c.Head.Children {
			if previewComp, ok := child.(*components.MJPreviewComponent); ok {
				if err := previewComp.Render(w); err != nil {
					return err
				}
			}
		}
	}

	// Write the body content (already rendered once above)
	if _, err := w.WriteString(bodyContent); err != nil {
		return err
	}
	if _, err := w.WriteString(`</body></html>`); err != nil {
		return err
	}

	return nil
}
