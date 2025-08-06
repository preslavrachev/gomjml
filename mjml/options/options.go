// Package options contains render options for MJML components
package options

import "sync"

// OutputFormat specifies the output format for rendering
type OutputFormat int

const (
	OutputHTML OutputFormat = iota // Default HTML rendering
	OutputMJML                     // MJML rendering
)

// FontTracker tracks font families used by components during rendering
type FontTracker struct {
	mu    sync.Mutex
	fonts map[string]bool // Set of unique font families
}

// NewFontTracker creates a new font tracker
func NewFontTracker() *FontTracker {
	return &FontTracker{
		fonts: make(map[string]bool),
	}
}

// AddFont adds a font family to the tracker if it maps to a Google Font
func (ft *FontTracker) AddFont(fontFamily string) {
	if fontFamily == "" {
		return
	}

	ft.mu.Lock()
	defer ft.mu.Unlock()
	ft.fonts[fontFamily] = true
}

// GetFonts returns all tracked font families as a slice
func (ft *FontTracker) GetFonts() []string {
	ft.mu.Lock()
	defer ft.mu.Unlock()

	fonts := make([]string, 0, len(ft.fonts))
	for font := range ft.fonts {
		fonts = append(fonts, font)
	}
	return fonts
}

// RenderOpts contains options for MJML rendering
type RenderOpts struct {
	DebugTags    bool         // Whether to include debug attributes in output
	InsideGroup  bool         // Whether the component is being rendered inside a group
	FontTracker  *FontTracker // Tracks fonts used during rendering
	OutputFormat OutputFormat // Output format (HTML=0 default, MJML=1)
}
