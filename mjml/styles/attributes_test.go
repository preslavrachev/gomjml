package styles

import (
	"testing"
)

func TestParsePixel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Pixel
		hasError bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
			hasError: false,
		},
		{
			name:     "valid pixel value with px",
			input:    "20px",
			expected: &Pixel{Value: 20},
			hasError: false,
		},
		{
			name:     "valid pixel value without px",
			input:    "15",
			expected: &Pixel{Value: 15},
			hasError: false,
		},
		{
			name:     "float pixel value",
			input:    "12.5px",
			expected: &Pixel{Value: 12.5},
			hasError: false,
		},
		{
			name:     "zero value",
			input:    "0px",
			expected: &Pixel{Value: 0},
			hasError: false,
		},
		{
			name:     "invalid value",
			input:    "invalid",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePixel(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}

			if tt.expected != nil && result == nil {
				t.Error("Expected result, got nil")
				return
			}

			if tt.expected != nil && result != nil {
				if result.Value != tt.expected.Value {
					t.Errorf("Expected value %f, got %f", tt.expected.Value, result.Value)
				}
			}
		})
	}
}

func TestPixelString(t *testing.T) {
	tests := []struct {
		name     string
		pixel    Pixel
		expected string
	}{
		{
			name:     "integer value",
			pixel:    Pixel{Value: 20},
			expected: "20px",
		},
		{
			name:     "float value rounded",
			pixel:    Pixel{Value: 12.7},
			expected: "13px",
		},
		{
			name:     "zero value",
			pixel:    Pixel{Value: 0},
			expected: "0px",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.pixel.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParseSpacing(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *Spacing
		hasError bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
			hasError: false,
		},
		{
			name:     "single value",
			input:    "10px",
			expected: &Spacing{Top: 10, Right: 10, Bottom: 10, Left: 10},
			hasError: false,
		},
		{
			name:     "two values",
			input:    "10px 20px",
			expected: &Spacing{Top: 10, Right: 20, Bottom: 10, Left: 20},
			hasError: false,
		},
		{
			name:     "four values",
			input:    "10px 20px 30px 40px",
			expected: &Spacing{Top: 10, Right: 20, Bottom: 30, Left: 40},
			hasError: false,
		},
		{
			name:     "values without px",
			input:    "5 10 15 20",
			expected: &Spacing{Top: 5, Right: 10, Bottom: 15, Left: 20},
			hasError: false,
		},
		{
			name:     "mixed px and no px",
			input:    "5px 10 15px 20",
			expected: &Spacing{Top: 5, Right: 10, Bottom: 15, Left: 20},
			hasError: false,
		},
		{
			name:     "three values (invalid)",
			input:    "10px 20px 30px",
			expected: nil,
			hasError: true,
		},
		{
			name:     "invalid value",
			input:    "invalid 20px",
			expected: nil,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSpacing(tt.input)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}

			if tt.expected != nil && result == nil {
				t.Error("Expected result, got nil")
				return
			}

			if tt.expected != nil && result != nil {
				if result.Top != tt.expected.Top ||
					result.Right != tt.expected.Right ||
					result.Bottom != tt.expected.Bottom ||
					result.Left != tt.expected.Left {
					t.Errorf("Expected %+v, got %+v", tt.expected, result)
				}
			}
		})
	}
}

func TestSpacingString(t *testing.T) {
	tests := []struct {
		name     string
		spacing  Spacing
		expected string
	}{
		{
			name:     "all same values",
			spacing:  Spacing{Top: 10, Right: 10, Bottom: 10, Left: 10},
			expected: "10px",
		},
		{
			name:     "different values",
			spacing:  Spacing{Top: 10, Right: 20, Bottom: 30, Left: 40},
			expected: "10px 20px 30px 40px",
		},
		{
			name:     "zero values",
			spacing:  Spacing{Top: 0, Right: 0, Bottom: 0, Left: 0},
			expected: "0px",
		},
		{
			name:     "mixed values",
			spacing:  Spacing{Top: 5, Right: 10, Bottom: 5, Left: 10},
			expected: "5px 10px",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.spacing.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParseColor(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *Color
		expectError bool
	}{
		{
			name:        "empty string",
			input:       "",
			expected:    nil,
			expectError: false,
		},
		{
			name:        "hex color with hash",
			input:       "#ff0000",
			expected:    &Color{Value: "#ff0000"},
			expectError: false,
		},
		{
			name:        "hex color without hash",
			input:       "ff0000",
			expected:    &Color{Value: "#ff0000"},
			expectError: false,
		},
		{
			name:        "short hex color",
			input:       "#f00",
			expected:    &Color{Value: "#f00"},
			expectError: false,
		},
		{
			name:        "named color",
			input:       "red",
			expected:    &Color{Value: "red"},
			expectError: false,
		},
		{
			name:        "rgb color",
			input:       "rgb(255, 0, 0)",
			expected:    &Color{Value: "rgb(255, 0, 0)"},
			expectError: false,
		},
		{
			name:        "rgba color",
			input:       "rgba(255, 0, 0, 0.5)",
			expected:    &Color{Value: "rgba(255, 0, 0, 0.5)"},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseColor(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}

			if tt.expected != nil && result == nil {
				t.Error("Expected result, got nil")
				return
			}

			if tt.expected != nil && result != nil {
				if result.Value != tt.expected.Value {
					t.Errorf("Expected value '%s', got '%s'", tt.expected.Value, result.Value)
				}
			}
		})
	}
}

func TestColorString(t *testing.T) {
	tests := []struct {
		name     string
		color    Color
		expected string
	}{
		{
			name:     "hex color",
			color:    Color{Value: "#ff0000"},
			expected: "#ff0000",
		},
		{
			name:     "named color",
			color:    Color{Value: "red"},
			expected: "red",
		},
		{
			name:     "rgb color",
			color:    Color{Value: "rgb(255, 0, 0)"},
			expected: "rgb(255, 0, 0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.color.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestParseNonEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: nil,
		},
		{
			name:     "non-empty string",
			input:    "test",
			expected: func() *string { s := "test"; return &s }(),
		},
		{
			name:     "whitespace string",
			input:    "   ",
			expected: func() *string { s := "   "; return &s }(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseNonEmpty(tt.input)

			if tt.expected == nil && result != nil {
				t.Errorf("Expected nil, got %v", result)
				return
			}

			if tt.expected != nil && result == nil {
				t.Error("Expected result, got nil")
				return
			}

			if tt.expected != nil && result != nil {
				if *result != *tt.expected {
					t.Errorf("Expected '%s', got '%s'", *tt.expected, *result)
				}
			}
		})
	}
}
