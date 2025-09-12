package styles

import "strings"

// ParseBorderWidth extracts the pixel width from a CSS border shorthand value.
// It returns 0 if the width cannot be determined.
func ParseBorderWidth(attr string) int {
	parts := strings.Fields(attr)
	if len(parts) > 0 {
		if px, err := ParsePixel(parts[0]); err == nil && px != nil {
			return int(px.Value)
		}
	}
	return 0
}
