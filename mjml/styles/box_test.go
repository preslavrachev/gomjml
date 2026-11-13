package styles

import "testing"

func TestParseLeadingInt(t *testing.T) {
	tests := []struct {
		in   string
		want int
		ok   bool
	}{
		{"25px", 25, true},
		{"5%", 5, true},
		{"12.9px", 12, true},
		{" -3px", -3, true},
		{"0", 0, true},
		{"auto", 0, false},
		{"", 0, false},
	}
	for _, tt := range tests {
		if got, ok := ParseLeadingInt(tt.in); got != tt.want || ok != tt.ok {
			t.Errorf("ParseLeadingInt(%q) = %d, %v; want %d, %v", tt.in, got, ok, tt.want, tt.ok)
		}
	}
}

func TestShorthandValue(t *testing.T) {
	tests := []struct {
		value, direction string
		want             int
	}{
		{"10px", "left", 10},
		{"0px 25px", "left", 25},
		{"0 25px", "right", 25},
		{"1px 2px 3px", "left", 2},
		{"1px 2px 3px", "bottom", 3},
		{"1px 2px 3px 4px", "left", 4},
		{"1px 2px 3px 4px", "right", 2},
		{"5%", "left", 5},
		{"1px   2px", "right", 2},
		{"1px 2px 3px 4px 5px", "left", 4},
	}
	for _, tt := range tests {
		if got := ShorthandValue(tt.value, tt.direction); got != tt.want {
			t.Errorf("ShorthandValue(%q, %q) = %d, want %d", tt.value, tt.direction, got, tt.want)
		}
	}
}

func TestBorderValue(t *testing.T) {
	tests := []struct {
		border string
		want   int
	}{
		{"2px dashed #ff0000", 2},
		{"solid 3px red", 3},
		{"1px solid #336699", 1},
		{"#000 solid 4px", 4},
		{"none", 0},
		{"0", 0},
		{"", 0},
	}
	for _, tt := range tests {
		if got := BorderValue(tt.border); got != tt.want {
			t.Errorf("BorderValue(%q) = %d, want %d", tt.border, got, tt.want)
		}
	}
}

func TestFormatJSNumber(t *testing.T) {
	for in, want := range map[float64]string{600: "600", 289.5: "289.5", 199.98: "199.98", -0.0: "0"} {
		if got := FormatJSNumber(in); got != want {
			t.Errorf("FormatJSNumber(%v) = %q, want %q", in, got, want)
		}
	}
}
