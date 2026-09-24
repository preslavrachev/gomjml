package styles

import (
	"regexp"
	"strconv"
	"strings"
)

// ParseLeadingInt parses value the way JavaScript's parseInt(value, 10) does:
// leading whitespace and an optional sign, then as many decimal digits as there
// are, ignoring what follows ("12.5px" is 12, "5%" is 5). ok is false where
// parseInt would return NaN.
func ParseLeadingInt(value string) (n int, ok bool) {
	s := strings.TrimLeft(value, " \t\n\r\f\v")
	neg := false
	if s != "" && (s[0] == '-' || s[0] == '+') {
		neg = s[0] == '-'
		s = s[1:]
	}
	end := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
	}
	if end == 0 {
		return 0, false
	}
	n, err := strconv.Atoi(s[:end])
	if err != nil {
		return 0, false
	}
	if neg {
		n = -n
	}
	return n, true
}

// ParseLeadingFloat parses value the way JavaScript's parseFloat does, taking
// the longest decimal prefix ("289.5px" is 289.5). ok is false for NaN.
func ParseLeadingFloat(value string) (f float64, ok bool) {
	s := strings.TrimLeft(value, " \t\n\r\f\v")
	end := 0
	if end < len(s) && (s[end] == '-' || s[end] == '+') {
		end++
	}
	digits := 0
	for end < len(s) && s[end] >= '0' && s[end] <= '9' {
		end++
		digits++
	}
	if end < len(s) && s[end] == '.' {
		end++
		for end < len(s) && s[end] >= '0' && s[end] <= '9' {
			end++
			digits++
		}
	}
	if digits == 0 {
		return 0, false
	}
	f, err := strconv.ParseFloat(strings.TrimSuffix(s[:end], "."), 64)
	return f, err == nil
}

// ShorthandValue returns one side of a 1–4 value CSS shorthand such as padding,
// as MJML's shorthandParser does: only the leading integer of each part counts,
// and a part that is missing or not a number counts as 0 (NaN for one value).
func ShorthandValue(value, direction string) int {
	parts := strings.Fields(value)
	if len(parts) > 4 {
		parts = parts[:4]
	}
	var index int
	switch len(parts) {
	case 2:
		index = map[string]int{"top": 0, "bottom": 0, "left": 1, "right": 1}[direction]
	case 3:
		index = map[string]int{"top": 0, "left": 1, "right": 1, "bottom": 2}[direction]
	case 4:
		index = map[string]int{"top": 0, "right": 1, "bottom": 2, "left": 3}[direction]
	default:
		n, _ := ParseLeadingInt(value)
		return n
	}
	n, _ := ParseLeadingInt(parts[index])
	return n
}

// BorderValue returns the width of a CSS border shorthand as MJML's
// borderParser does: the first run of digits at the start or after a space
// ("2px solid red" and "solid 2px red" are 2, "none" is 0).
func BorderValue(border string) int {
	for i := 0; i < len(border); i++ {
		if border[i] < '0' || border[i] > '9' || (i > 0 && border[i-1] != ' ') {
			continue
		}
		end := i
		for end < len(border) && border[end] >= '0' && border[end] <= '9' {
			end++
		}
		n, err := strconv.Atoi(border[i:end])
		if err != nil {
			return 0
		}
		return n
	}
	return 0
}

// FormatJSNumber formats f as JavaScript's String(number) does for the values
// widths take: the shortest decimal that round-trips (289.5, 199.98, 200).
func FormatJSNumber(f float64) string {
	if f == 0 {
		return "0" // String(-0) is "0"
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// WidthUnit returns the unit MJML's widthParser reads from width, "px" when
// there is none: what follows the last run of digits, dots and commas.
func WidthUnit(width string) string {
	if m := widthUnitPattern.FindStringSubmatch(width); m != nil && m[1] != "" {
		return m[1]
	}
	return "px"
}

// widthUnitPattern is widthParser's unitRegex.
var widthUnitPattern = regexp.MustCompile(`[\d.,]*(\D*)$`)
