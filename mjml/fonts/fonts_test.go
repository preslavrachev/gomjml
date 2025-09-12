package fonts

import (
	"regexp"
	"strings"
	"testing"
)

func TestBuildFontsTags_SingleFontFormatting(t *testing.T) {
	url := GoogleFontsMapping["Ubuntu"]
	urls := []string{url}
	out := BuildFontsTags(urls)

	// Required substrings
	required := []string{
		"<!--[if !mso]><!-->",
		url,
		"<style type=\"text/css\">",
		"@import url(" + url + ");",
		"<!--<![endif]-->",
	}
	for _, sub := range required {
		if !strings.Contains(out, sub) {
			t.Fatalf("output missing required substring %q:\n%s", sub, out)
		}
	}

	// Order constraints (semantic, not formatting)
	openIdx := strings.Index(out, "<!--[if !mso]><!-->")
	linkIdx := strings.Index(out, `<link href="`+url+`"`)
	styleIdx := strings.Index(out, `<style type="text/css">`)
	importIdx := strings.Index(out, `@import url(`+url+`);`)
	closeIdx := strings.Index(out, "<!--<![endif]-->")

	if !(openIdx < linkIdx && linkIdx < styleIdx && styleIdx < importIdx && importIdx < closeIdx) {
		t.Errorf("Incorrect semantic order: open=%d link=%d style=%d import=%d close=%d", openIdx, linkIdx, styleIdx, importIdx, closeIdx)
	}

	// Single occurrences
	if c := strings.Count(out, "<link "); c != 1 {
		t.Errorf("expected 1 link tag, got %d", c)
	}
	if c := strings.Count(out, "@import url("); c != 1 {
		t.Errorf("expected 1 @import, got %d", c)
	}

	// Verify import is inside the style block
	styleStart := styleIdx + len(`<style type="text/css">`)
	styleEnd := strings.Index(out[styleStart:], "</style>")
	if styleEnd < 0 {
		t.Fatalf("style closing tag not found; output=%s", out)
	}
	styleContent := out[styleStart : styleStart+styleEnd]
	if !strings.Contains(styleContent, "@import url("+url+");") {
		t.Errorf("style block missing import; styleContent=%q", styleContent)
	}

	// Optional: If you intentionally require no newlines (document why)
	// if strings.Contains(out, "\n") {
	// 	t.Errorf("Unexpected newline in output (format contract): %q", out)
	// }
}

func TestBuildFontsTags_MultipleFonts(t *testing.T) {
	urls := []string{
		GoogleFontsMapping["Ubuntu"],
		GoogleFontsMapping["Roboto"],
	}
	out := BuildFontsTags(urls)

	// Both links present once
	for _, u := range urls {
		if strings.Count(out, u) < 2 { // appears in link + import
			t.Fatalf("expected font URL %s in both link and import at least twice aggregate, output=%s", u, out)
		}
		if strings.Count(out, `<link href="`+u+`"`) != 1 {
			t.Errorf("expected exactly 1 link tag for %s", u)
		}
		if strings.Count(out, "@import url("+u+")") != 1 {
			t.Errorf("expected exactly 1 @import for %s", u)
		}
	}

	// Order: all link tags must precede style block
	styleIdx := strings.Index(out, "<style type=\"text/css\">")
	if styleIdx < 0 {
		t.Fatalf("style tag not found")
	}
	linkTagPattern := regexp.MustCompile(`<link href="https://fonts\.googleapis\.com/css\?family=[^"]+" rel="stylesheet" type="text/css">`)
	locs := linkTagPattern.FindAllStringIndex(out, -1)
	for _, loc := range locs {
		if loc[0] > styleIdx {
			t.Errorf("link tag appears after style block at index %d", loc[0])
		}
	}

	// No duplicates in link section (quick check)
	if strings.Count(out[:styleIdx], "<link ") != len(urls) {
		t.Errorf("unexpected number of link tags before style; got %d want %d", strings.Count(out[:styleIdx], "<link "), len(urls))
	}
}
