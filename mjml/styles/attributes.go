package styles

import (
	"fmt"
	"strconv"
	"strings"
)

// Pixel represents a pixel value for CSS properties.
// It provides type-safe handling of pixel-based measurements commonly used in CSS.
type Pixel struct {
	Value float64 // The numeric pixel value (e.g., 20 for "20px")
}

// String returns the CSS representation of the pixel value.
// The value is formatted as an integer with "px" suffix.
//
// Example:
//
//	p := Pixel{Value: 20.5}
//	fmt.Println(p.String()) // "21px" (rounded to nearest integer)
func (p Pixel) String() string {
	return fmt.Sprintf("%.0fpx", p.Value)
}

// Spacing represents CSS spacing values (margin, padding) with separate values for each side.
// This follows the CSS box model with Top, Right, Bottom, Left values.
type Spacing struct {
	Top, Right, Bottom, Left float64 // Pixel values for each side of the box
}

// String returns the CSS representation of the spacing value.
// It automatically optimizes the output format:
// - If all sides are equal: "10px"
// - If top/bottom and left/right are equal: "10px 20px"
// - Otherwise: "10px 20px 30px 40px"
//
// Example:
//
//	s := Spacing{Top: 10, Right: 20, Bottom: 10, Left: 20}
//	fmt.Println(s.String()) // "10px 20px"
func (s Spacing) String() string {
	if s.Top == s.Right && s.Right == s.Bottom && s.Bottom == s.Left {
		return fmt.Sprintf("%.0fpx", s.Top)
	}
	if s.Top == s.Bottom && s.Right == s.Left {
		return fmt.Sprintf("%.0fpx %.0fpx", s.Top, s.Right)
	}
	return fmt.Sprintf("%.0fpx %.0fpx %.0fpx %.0fpx", s.Top, s.Right, s.Bottom, s.Left)
}

// ParsePixel parses a string value into a Pixel struct.
// It accepts values with or without the "px" suffix and returns nil for empty strings.
//
// Supported formats:
//   - "20" -> Pixel{Value: 20}
//   - "20px" -> Pixel{Value: 20}
//   - "" -> nil (no error)
//
// Returns an error for invalid numeric values.
//
// Example:
//
//	pixel, err := ParsePixel("20px")
//	if err == nil {
//	    fmt.Println(pixel.String()) // "20px"
//	}
func ParsePixel(value string) (*Pixel, error) {
	if value == "" {
		return nil, nil
	}

	cleaned := strings.TrimSuffix(value, "px")
	val, err := strconv.ParseFloat(cleaned, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid pixel value: %s", value)
	}

	return &Pixel{Value: val}, nil
}

// ParseSpacing parses a CSS spacing string into a Spacing struct.
// It supports standard CSS spacing formats and returns nil for empty strings.
//
// Supported formats:
//   - "10px" -> all sides = 10
//   - "10px 20px" -> top/bottom = 10, left/right = 20
//   - "10px 20px 30px 40px" -> top=10, right=20, bottom=30, left=40
//
// The 3-value format ("10px 20px 30px") is not currently supported and will return an error.
//
// Returns an error for invalid formats or unparseable pixel values.
//
// Example:
//
//	spacing, err := ParseSpacing("10px 20px")
//	if err == nil {
//	    fmt.Println(spacing.String()) // "10px 20px"
//	}
func ParseSpacing(value string) (*Spacing, error) {
	if value == "" {
		return nil, nil
	}

	parts := strings.Fields(value)
	var spacing Spacing

	switch len(parts) {
	case 1:
		val, err := ParsePixel(parts[0])
		if err != nil {
			return nil, err
		}
		spacing = Spacing{val.Value, val.Value, val.Value, val.Value}
	case 2:
		topBottom, err := ParsePixel(parts[0])
		if err != nil {
			return nil, err
		}
		leftRight, err := ParsePixel(parts[1])
		if err != nil {
			return nil, err
		}
		spacing = Spacing{topBottom.Value, leftRight.Value, topBottom.Value, leftRight.Value}
	case 4:
		top, err := ParsePixel(parts[0])
		if err != nil {
			return nil, err
		}
		right, err := ParsePixel(parts[1])
		if err != nil {
			return nil, err
		}
		bottom, err := ParsePixel(parts[2])
		if err != nil {
			return nil, err
		}
		left, err := ParsePixel(parts[3])
		if err != nil {
			return nil, err
		}
		spacing = Spacing{top.Value, right.Value, bottom.Value, left.Value}
	default:
		return nil, fmt.Errorf("invalid spacing format: %s", value)
	}

	return &spacing, nil
}

// Color represents a CSS color value with validation and normalization.
// It handles various CSS color formats including hex, named colors, rgb, rgba, etc.
type Color struct {
	Value string // The CSS color value (e.g., "#ff0000", "red", "rgb(255,0,0)")
}

// String returns the CSS representation of the color value.
// The value is returned as-is since colors can have various valid formats.
//
// Example:
//
//	c := Color{Value: "#ff0000"}
//	fmt.Println(c.String()) // "#ff0000"
func (c Color) String() string {
	return c.Value
}

// ParseColor parses a string value into a Color struct.
// It accepts various CSS color formats and normalizes hex colors without hash prefix.
//
// Supported formats:
//   - "#ff0000" -> Color{Value: "#ff0000"}
//   - "ff0000" -> Color{Value: "#ff0000"} (adds missing #)
//   - "red" -> Color{Value: "red"}
//   - "rgb(255, 0, 0)" -> Color{Value: "rgb(255, 0, 0)"}
//   - "" -> nil (no error)
//
// Example:
//
//	color, err := ParseColor("#ff0000")
//	if err == nil {
//	    fmt.Println(color.String()) // "#ff0000"
//	}
func ParseColor(value string) (*Color, error) {
	if value == "" {
		return nil, nil
	}

	// Normalize hex colors without # prefix
	if len(value) == 6 && isHexColor(value) {
		value = "#" + value
	}

	return &Color{Value: value}, nil
}

// isHexColor checks if a string contains only valid hex color characters
func isHexColor(s string) bool {
	if len(s) != 3 && len(s) != 6 {
		return false
	}
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}

// Size represents CSS size values (width, height) that can be in pixels or percentages.
// This matches the MRML Size enum structure for compatibility.
type Size struct {
	value   float64
	isPixel bool // true for pixels, false for percentages
}

// NewPixelSize creates a new pixel-based Size
func NewPixelSize(value float64) Size {
	return Size{value: value, isPixel: true}
}

// NewPercentSize creates a new percentage-based Size
func NewPercentSize(value float64) Size {
	return Size{value: value, isPixel: false}
}

// Value returns the numeric value of the size
func (s Size) Value() float64 {
	return s.value
}

// IsPixel returns true if this is a pixel-based size
func (s Size) IsPixel() bool {
	return s.isPixel
}

// IsPercent returns true if this is a percentage-based size
func (s Size) IsPercent() bool {
	return !s.isPixel
}

// String returns the CSS representation of the size
func (s Size) String() string {
	if s.isPixel {
		return fmt.Sprintf("%.0fpx", s.value)
	}
	formatted := strconv.FormatFloat(s.value, 'g', -1, 64)
	if formatted == "" {
		formatted = "0"
	}
	return formatted + "%"
}

// ParseSize parses a string value into a Size struct.
// It supports both pixel and percentage formats.
//
// Supported formats:
//   - "20px" -> Size{value: 20, isPixel: true}
//   - "20" -> Size{value: 20, isPixel: true} (assumes pixels)
//   - "50%" -> Size{value: 50, isPixel: false}
//
// Returns an error for invalid formats.
func ParseSize(value string) (Size, error) {
	if value == "" {
		return Size{}, fmt.Errorf("empty size value")
	}

	// Check for percentage
	if strings.HasSuffix(value, "%") {
		numStr := strings.TrimSuffix(value, "%")
		val, err := strconv.ParseFloat(numStr, 64)
		if err != nil {
			return Size{}, fmt.Errorf("invalid percentage value: %s", value)
		}
		return NewPercentSize(val), nil
	}

	// Check for pixels (with or without px suffix)
	numStr := strings.TrimSuffix(value, "px")
	val, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return Size{}, fmt.Errorf("invalid pixel value: %s", value)
	}
	return NewPixelSize(val), nil
}

// parseNonEmpty is a utility function that returns a pointer to a string if it's non-empty,
// or nil if the string is empty. This is used throughout the styles package for conditional
// CSS property application.
//
// This function enables the "maybe add" pattern where CSS properties are only added
// if their values are actually present.
//
// Example:
//
//	var color string = getColor() // might be empty
//	tag.MaybeAddStyle("color", parseNonEmpty(color)) // only adds if color is not empty
func parseNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
