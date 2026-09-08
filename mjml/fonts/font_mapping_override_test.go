package fonts

import (
	"slices"
	"testing"
)

// TestGetGoogleFontURLUsesMappingOverride verifies that GetGoogleFontURL
// reads its URL from the live GoogleFontsMapping value at call time, so a
// caller overriding an entry (e.g. pointing "Roboto" at a self-hosted mirror)
// changes what GetGoogleFontURL returns, rather than being silently ignored.
func TestGetGoogleFontURLUsesMappingOverride(t *testing.T) {
	//GIVEN: GoogleFontsMapping's "Roboto" entry overridden with a custom URL.
	original := GoogleFontsMapping["Roboto"]
	t.Cleanup(func() { GoogleFontsMapping["Roboto"] = original })
	const customURL = "https://fonts.example.com/roboto.css"
	GoogleFontsMapping["Roboto"] = customURL

	//WHEN: GetGoogleFontURL resolves a font family containing "Roboto".
	got := GetGoogleFontURL("Roboto, Arial, sans-serif")

	//THEN: it returns the overridden URL, not the original built-in one.
	if got != customURL {
		t.Errorf("GetGoogleFontURL() = %q, want overridden URL %q", got, customURL)
	}
}

// TestConvertFontFamiliesToURLsUsesAddedMappingEntry verifies that
// ConvertFontFamiliesToURLs picks up a font family added to GoogleFontsMapping
// at runtime, confirming the map remains the single source of truth for both
// functions rather than a second, independent registry.
func TestConvertFontFamiliesToURLsUsesAddedMappingEntry(t *testing.T) {
	//GIVEN: a new font family added to GoogleFontsMapping that isn't one of the built-in entries.
	const name = "Comic Neue"
	const url = "https://fonts.example.com/comic-neue.css"
	if _, exists := GoogleFontsMapping[name]; exists {
		t.Fatalf("test fixture %q unexpectedly already present in GoogleFontsMapping", name)
	}
	t.Cleanup(func() { delete(GoogleFontsMapping, name) })
	GoogleFontsMapping[name] = url

	//WHEN: ConvertFontFamiliesToURLs resolves a font stack using the added family.
	got := ConvertFontFamiliesToURLs([]string{"Comic Neue, cursive"})

	//THEN: the added entry's URL is present in the result.
	if !slices.Contains(got, url) {
		t.Errorf("ConvertFontFamiliesToURLs() = %v, want it to include added entry URL %q", got, url)
	}
}

// TestConvertFontFamiliesToURLsDeduplicatesAliasedURLs verifies that when two
// mapping names share the same URL (e.g. GoogleFontsMapping["Alias"] =
// GoogleFontsMapping["Roboto"]), and a font family matches both names,
// ConvertFontFamiliesToURLs emits that URL only once rather than once per
// matching name.
func TestConvertFontFamiliesToURLsDeduplicatesAliasedURLs(t *testing.T) {
	//GIVEN: "Roboto Alias" mapped to the same URL as the existing "Roboto" entry.
	original := GoogleFontsMapping["Roboto"]
	t.Cleanup(func() { delete(GoogleFontsMapping, "Roboto Alias") })
	GoogleFontsMapping["Roboto Alias"] = original

	//WHEN: a single font family matches both "Roboto" and "Roboto Alias".
	got := ConvertFontFamiliesToURLs([]string{"Roboto Alias, Arial, sans-serif"})

	//THEN: the shared URL is emitted exactly once.
	count := 0
	for _, u := range got {
		if u == original {
			count++
		}
	}
	if count != 1 {
		t.Errorf("ConvertFontFamiliesToURLs() = %v, want URL %q exactly once, got %d occurrences", got, original, count)
	}
}

// TestConvertFontFamiliesToURLsSkipsEmptyURLs verifies that a mapping entry
// whose URL is empty never produces an empty entry in the result, matching
// the historical behavior of skipping lookups that resolve to nothing.
func TestConvertFontFamiliesToURLsSkipsEmptyURLs(t *testing.T) {
	//GIVEN: a mapping entry with an empty URL value.
	const name = "Empty Font"
	if _, exists := GoogleFontsMapping[name]; exists {
		t.Fatalf("test fixture %q unexpectedly already present in GoogleFontsMapping", name)
	}
	t.Cleanup(func() { delete(GoogleFontsMapping, name) })
	GoogleFontsMapping[name] = ""

	//WHEN: a font family matches only that empty-URL entry.
	got := ConvertFontFamiliesToURLs([]string{"Empty Font, cursive"})

	//THEN: no empty string is present in the result.
	if slices.Contains(got, "") {
		t.Errorf("ConvertFontFamiliesToURLs() = %v, want no empty URL entries", got)
	}
}
