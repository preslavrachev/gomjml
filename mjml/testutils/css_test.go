package testutils

import (
	"testing"
)

func TestStylesEqual(t *testing.T) {
	tests := []struct {
		name   string
		style1 string
		style2 string
		want   bool
	}{
		{
			name:   "identical strings",
			style1: "color: red; font-size: 12px;",
			style2: "color: red; font-size: 12px;",
			want:   true,
		},
		{
			name:   "different order same properties",
			style1: "color: red; font-size: 12px;",
			style2: "font-size: 12px; color: red;",
			want:   true,
		},
		{
			name:   "different whitespace",
			style1: "color:red;font-size:12px;",
			style2: "color: red; font-size: 12px;",
			want:   true,
		},
		{
			name:   "extra whitespace and semicolons",
			style1: "color: red;  font-size: 12px;;",
			style2: "font-size: 12px; color: red",
			want:   true,
		},
		{
			name:   "empty strings",
			style1: "",
			style2: "",
			want:   true,
		},
		{
			name:   "one empty one with content",
			style1: "",
			style2: "color: red;",
			want:   false,
		},
		{
			name:   "different values",
			style1: "color: red;",
			style2: "color: blue;",
			want:   false,
		},
		{
			name:   "different properties",
			style1: "color: red;",
			style2: "background: red;",
			want:   false,
		},
		{
			name:   "subset properties",
			style1: "color: red; font-size: 12px;",
			style2: "color: red;",
			want:   false,
		},
		{
			name:   "complex real-world example",
			style1: "font-size:0px;padding:20px;word-break:break-word;",
			style2: "font-size:0px;word-break:break-word;padding:20px;",
			want:   true,
		},
		{
			name:   "border radius different order",
			style1: "border:0;border-radius:10px;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;",
			style2: "border:0;display:block;outline:none;text-decoration:none;height:auto;width:100%;font-size:13px;border-radius:10px;",
			want:   true,
		},
		{
			name:   "malformed CSS ignored",
			style1: "color: red; invalid-no-colon; font-size: 12px;",
			style2: "color: red; font-size: 12px;",
			want:   true,
		},
		{
			name:   "properties with multiple colons",
			style1: "background: url('http://example.com/image.png');",
			style2: "background: url('http://example.com/image.png');",
			want:   true,
		},
		{
			name:   "case sensitive property names",
			style1: "Color: red;",
			style2: "color: red;",
			want:   false,
		},
		{
			name:   "case sensitive property values",
			style1: "color: RED;",
			style2: "color: red;",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StylesEqual(tt.style1, tt.style2)
			if got != tt.want {
				t.Errorf("StylesEqual(%q, %q) = %v, want %v", tt.style1, tt.style2, got, tt.want)
			}
		})
	}
}

func TestParseStyleProperties(t *testing.T) {
	tests := []struct {
		name  string
		style string
		want  map[string]string
	}{
		{
			name:  "empty string",
			style: "",
			want:  map[string]string{},
		},
		{
			name:  "single property",
			style: "color: red;",
			want:  map[string]string{"color": "red"},
		},
		{
			name:  "multiple properties",
			style: "color: red; font-size: 12px;",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "no trailing semicolon",
			style: "color: red",
			want:  map[string]string{"color": "red"},
		},
		{
			name:  "extra whitespace",
			style: "  color : red ;  font-size : 12px  ;  ",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "empty declarations",
			style: "color: red;; ; font-size: 12px;",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "malformed declarations ignored",
			style: "color: red; invalid-no-colon; font-size: 12px;",
			want:  map[string]string{"color": "red", "font-size": "12px"},
		},
		{
			name:  "property with multiple colons",
			style: "background: url('http://example.com/image.png');",
			want:  map[string]string{"background": "url('http://example.com/image.png')"},
		},
		{
			name:  "empty property or value ignored",
			style: ": red; color:; color: blue;",
			want:  map[string]string{"color": "blue"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStyleProperties(tt.style)
			if len(got) != len(tt.want) {
				t.Errorf("parseStyleProperties(%q) returned %d properties, want %d", tt.style, len(got), len(tt.want))
			}
			for prop, expectedValue := range tt.want {
				if actualValue, exists := got[prop]; !exists {
					t.Errorf("parseStyleProperties(%q) missing property %q", tt.style, prop)
				} else if actualValue != expectedValue {
					t.Errorf("parseStyleProperties(%q) property %q = %q, want %q", tt.style, prop, actualValue, expectedValue)
				}
			}
			for prop := range got {
				if _, exists := tt.want[prop]; !exists {
					t.Errorf("parseStyleProperties(%q) unexpected property %q = %q", tt.style, prop, got[prop])
				}
			}
		})
	}
}

// Benchmark tests to ensure performance is reasonable
func BenchmarkStylesEqual(b *testing.B) {
	style1 := "font-size:0px;padding:20px;word-break:break-word;color:red;background:white;border:1px solid black;"
	style2 := "color:red;font-size:0px;border:1px solid black;word-break:break-word;background:white;padding:20px;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		StylesEqual(style1, style2)
	}
}

func BenchmarkParseStyleProperties(b *testing.B) {
	style := "font-size:0px;padding:20px;word-break:break-word;color:red;background:white;border:1px solid black;"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		parseStyleProperties(style)
	}
}
