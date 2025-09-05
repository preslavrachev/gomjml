package components

import (
	"html"
	"regexp"
	"strconv"
	"strings"
)

var percentRe = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)%$`)

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

// mapKeywordOrPercentToVMLDecimal converts CSS keyword/percent to decimal string
func mapKeywordOrPercentToVMLDecimal(v string, horiz bool) string {
	switch v {
	case "left", "top":
		return "0"
	case "center":
		return "0.5"
	case "right", "bottom":
		return "1"
	}
	if m := percentRe.FindStringSubmatch(v); m != nil {
		if f, err := strconv.ParseFloat(m[1], 64); err == nil {
			return strconv.FormatFloat(f/100.0, 'f', -1, 64)
		}
	}
	// Length values not fully supported; fallback center
	return "0.5"
}

func buildBackgroundShorthand(color, url, posX, posY, size, repeat string) string {
	// Similar to MJML: color + url('...') + posX posY / size + repeat
	parts := []string{}
	if color != "" {
		parts = append(parts, color)
	}
	if url != "" {
		parts = append(parts, "url('"+url+"')")
	}
	if posX != "" || posY != "" {
		parts = append(parts, posX+" "+posY)
	}
	if size != "" && size != "auto" {
		parts = append(parts, "/", size)
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

func computeVMLPosition(posX, posY, size string) (originX, originY, posValX, posValY string) {
	// Map keywords to decimal values
	mapDecimal := func(v string, horiz bool) string {
		switch v {
		case "left", "top":
			return "0"
		case "center":
			return "0.5"
		case "right", "bottom":
			return "1"
		}
		// percent (e.g. 30%) => 0.3
		if strings.HasSuffix(v, "%") {
			p := strings.TrimSuffix(v, "%")
			if f, err := strconv.ParseFloat(p, 64); err == nil {
				return strconv.FormatFloat(f/100.0, 'f', -1, 64)
			}
		}
		// default
		return "0.5"
	}

	// Get base decimal values
	decX := mapDecimal(posX, true)
	decY := mapDecimal(posY, false)

	// For cover sizing, MJML uses special VML positioning logic
	if size == "cover" {
		// Convert position to VML values: center=0, top=-0.5
		// This matches MJML's VML generation for cover backgrounds
		vmlX := "0"    // center -> 0
		vmlY := "-0.5" // top -> -0.5

		if posX == "left" {
			vmlX = "-0.5"
		} else if posX == "right" {
			vmlX = "0.5"
		}

		if posY == "center" {
			vmlY = "0"
		} else if posY == "bottom" {
			vmlY = "0.5"
		}

		return vmlX, vmlY, vmlX, vmlY
	}

	// For other sizes, use standard mapping
	return decX, decY, decX, decY
}

func htmlEscape(s string) string {
	return html.EscapeString(s)
}
