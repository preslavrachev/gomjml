package components

import (
	"html"
	"strconv"
	"strings"
)

// parseBackgroundPosition converts CSS keywords/percent/length into canonical (xKeyword, yKeyword)
func parseBackgroundPosition(raw string) (string, string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "top", "center" // fallback consistent with MJML default "top center"
	}
	toks := strings.Fields(raw)
	if len(toks) == 1 {
		v := toks[0]
		if v == "top" || v == "bottom" {
			return "center", v
		}
		return v, "center"
	}
	if len(toks) >= 2 {
		// Determine which is x vs y similar to MJML JS logic
		v1, v2 := toks[0], toks[1]
		if isVertical(v1) || (v1 == "center" && isHorizontal(v2)) {
			return v2, v1
		}
		return v1, v2
	}
	return "center", "center"
}

func isHorizontal(v string) bool {
	switch v {
	case "left", "right":
		return true
	}
	return false
}

func isVertical(v string) bool {
	switch v {
	case "top", "bottom":
		return true
	}
	return false
}

// overridePosition applies background-position-x / -y overrides
func overridePosition(x, y, posXOverride, posYOverride string) (string, string) {
	if posXOverride != "" {
		x = posXOverride
	}
	if posYOverride != "" {
		y = posYOverride
	}
	return x, y
}

func buildBackgroundShorthand(color, url, posX, posY, size, repeat string) string {
	// Similar to MJML: color + url('...') + posX posY / size + repeat
	parts := []string{}
	// Only include color if it's explicitly provided (including "transparent")
	if color != "" {
		parts = append(parts, color)
	}
	if url != "" {
		parts = append(parts, "url('"+htmlEscape(url)+"')")
	}
	if posX != "" || posY != "" {
		position := posX + " " + posY
		// Include size in the position part with slash delimiter if we have a size
		if size != "" {
			position += " / " + size
		}
		parts = append(parts, position)
	}
	if repeat != "" {
		parts = append(parts, repeat)
	}
	return strings.Join(parts, " ")
}

func computeVMLType(repeat, size string) string {
	if repeat == "no-repeat" && size != "auto" {
		return "frame"
	}
	return "tile"
}

func computeVMLSize(size string) (sizeAttr string, aspect string) {
	switch size {
	case "cover":
		return `size="1,1"`, "atleast"
	case "contain":
		return `size="1,1"`, "atmost"
	case "", "auto":
		return "", ""
	}
	parts := strings.Fields(size)
	if len(parts) == 1 {
		return `size="` + parts[0] + `"`, "atmost"
	}
	if len(parts) >= 2 {
		return `size="` + parts[0] + `,` + parts[1] + `"`, ""
	}
	return "", ""
}

// AIDEV-NOTE: VML positioning depends on background-repeat mode - see docs/vml-background-positioning.md
func computeVMLPosition(posX, posY, _ string, repeat string) (originX, originY, posValX, posValY string) {
	// VML positioning depends on background-repeat mode:
	// - repeat (tile): Direct mapping - center → 0.5
	// - no-repeat (frame): Shifted mapping - center → 0 (0.5 - 0.5 = 0)
	isFrameMode := repeat == "no-repeat"

	mapDecimal := func(v string, _ bool) string {
		var baseVal float64
		switch v {
		case "left", "top":
			baseVal = 0
		case "center":
			baseVal = 0.5
		case "right", "bottom":
			baseVal = 1
		default:
			// percent (e.g. 30%) => 0.3
			if strings.HasSuffix(v, "%") {
				p := strings.TrimSuffix(v, "%")
				if f, err := strconv.ParseFloat(p, 64); err == nil {
					baseVal = f / 100.0
				} else {
					baseVal = 0.5 // default
				}
			} else {
				baseVal = 0.5 // default
			}
		}

		// Apply frame mode adjustment if needed
		if isFrameMode {
			baseVal = baseVal - 0.5
		}

		return strconv.FormatFloat(baseVal, 'f', -1, 64)
	}

	// Get decimal values with repeat-mode consideration
	decX := mapDecimal(posX, true)
	decY := mapDecimal(posY, false)

	return decX, decY, decX, decY
}

func htmlEscape(s string) string {
	return html.EscapeString(s)
}
