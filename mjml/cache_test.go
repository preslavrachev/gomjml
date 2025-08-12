package mjml

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// helper to clear cache and stop cleanup between tests
func resetASTCache() {
	astCache.Range(func(key, _ interface{}) bool {
		astCache.Delete(key)
		return true
	})
	StopASTCacheCleanup()
}

func TestCachingDisabledByDefault(t *testing.T) {
	resetASTCache()
	defer resetASTCache()

	var calls int32
	origParse := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		atomic.AddInt32(&calls, 1)
		return origParse(s)
	}
	defer func() { ParseMJML = origParse }()

	tpl := `<mjml><mj-body><mj-section><mj-column><mj-text>hi</mj-text></mj-column></mj-section></mj-body></mjml>`

	if _, err := Render(tpl); err != nil {
		t.Fatalf("render1: %v", err)
	}
	if _, err := Render(tpl); err != nil {
		t.Fatalf("render2: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 parses, got %d", calls)
	}

	entries := 0
	astCache.Range(func(_, _ interface{}) bool { entries++; return true })
	if entries != 0 {
		t.Fatalf("expected cache to remain empty, got %d entries", entries)
	}
}

func TestCachingStoresAndReusesAST(t *testing.T) {
	resetASTCache()
	defer resetASTCache()

	var calls int32
	origParse := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		atomic.AddInt32(&calls, 1)
		return origParse(s)
	}
	defer func() { ParseMJML = origParse }()

	tpl := `<mjml><mj-body><mj-section><mj-column><mj-text>hi</mj-text></mj-column></mj-section></mj-body></mjml>`

	r1, err := RenderWithAST(tpl, WithCache(true))
	if err != nil {
		t.Fatalf("render1: %v", err)
	}
	r2, err := RenderWithAST(tpl, WithCache(true))
	if err != nil {
		t.Fatalf("render2: %v", err)
	}

	if calls != 1 {
		t.Fatalf("expected 1 parse, got %d", calls)
	}
	if r1.AST != r2.AST {
		t.Fatalf("expected cached AST to be reused")
	}

	entries := 0
	astCache.Range(func(_, _ interface{}) bool { entries++; return true })
	if entries != 1 {
		t.Fatalf("expected 1 cache entry, got %d", entries)
	}
}

func TestCacheExpiration(t *testing.T) {
	resetASTCache()
	defer resetASTCache()

	origTTL := astCacheTTL
	astCacheTTL = 50 * time.Millisecond
	defer func() { astCacheTTL = origTTL }()

	var calls int32
	origParse := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		atomic.AddInt32(&calls, 1)
		return origParse(s)
	}
	defer func() { ParseMJML = origParse }()

	tpl := `<mjml><mj-body><mj-section><mj-column><mj-text>hi</mj-text></mj-column></mj-section></mj-body></mjml>`

	r1, err := RenderWithAST(tpl, WithCache(true))
	if err != nil {
		t.Fatalf("render1: %v", err)
	}
	time.Sleep(astCacheTTL + 10*time.Millisecond)
	r2, err := RenderWithAST(tpl, WithCache(true))
	if err != nil {
		t.Fatalf("render2: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 parses due to expiration, got %d", calls)
	}
	if r1.AST == r2.AST {
		t.Fatalf("expected new AST after expiration")
	}
}

func TestCacheConcurrentParsingSingleParse(t *testing.T) {
	resetASTCache()
	defer resetASTCache()

	var calls int32
	origParse := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		atomic.AddInt32(&calls, 1)
		time.Sleep(50 * time.Millisecond)
		return origParse(s)
	}
	defer func() { ParseMJML = origParse }()

	tpl := `<mjml><mj-body><mj-section><mj-column><mj-text>hi</mj-text></mj-column></mj-section></mj-body></mjml>`

	var wg sync.WaitGroup
	start := make(chan struct{})
	n := 5
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			if _, err := parseAST(tpl, true); err != nil {
				t.Errorf("parse: %v", err)
			}
		}()
	}
	close(start)
	wg.Wait()

	if calls != 1 {
		t.Fatalf("expected single parse, got %d", calls)
	}
}

func TestStopASTCacheCleanup(t *testing.T) {
	resetASTCache()
	defer resetASTCache()

	tpl := `<mjml><mj-body><mj-section><mj-column><mj-text>hi</mj-text></mj-column></mj-section></mj-body></mjml>`

	if _, err := RenderWithAST(tpl, WithCache(true)); err != nil {
		t.Fatalf("render: %v", err)
	}
	if cleanupCancel == nil {
		t.Fatalf("expected cleanup goroutine to start")
	}

	StopASTCacheCleanup()
	if cleanupCancel != nil {
		t.Fatalf("expected cleanup goroutine to stop")
	}
}

func TestCacheSeparateTemplates(t *testing.T) {
	resetASTCache()
	defer resetASTCache()

	var calls int32
	origParse := ParseMJML
	ParseMJML = func(s string) (*MJMLNode, error) {
		atomic.AddInt32(&calls, 1)
		return origParse(s)
	}
	defer func() { ParseMJML = origParse }()

	tpl1 := `<mjml><mj-body><mj-section><mj-column><mj-text>one</mj-text></mj-column></mj-section></mj-body></mjml>`
	tpl2 := `<mjml><mj-body><mj-section><mj-column><mj-text>two</mj-text></mj-column></mj-section></mj-body></mjml>`

	if _, err := RenderWithAST(tpl1, WithCache(true)); err != nil {
		t.Fatalf("render tpl1: %v", err)
	}
	if _, err := RenderWithAST(tpl2, WithCache(true)); err != nil {
		t.Fatalf("render tpl2: %v", err)
	}
	if _, err := RenderWithAST(tpl1, WithCache(true)); err != nil {
		t.Fatalf("render tpl1 again: %v", err)
	}

	if calls != 2 {
		t.Fatalf("expected 2 parses for different templates, got %d", calls)
	}

	entries := 0
	astCache.Range(func(_, _ interface{}) bool { entries++; return true })
	if entries != 2 {
		t.Fatalf("expected 2 cache entries, got %d", entries)
	}
}
