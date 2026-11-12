package html

import (
	"errors"
	"strings"
	"testing"
)

func TestNewHTMLTag(t *testing.T) {
	tag := NewHTMLTag("div")

	if tag.name != "div" {
		t.Errorf("Expected tag name 'div', got '%s'", tag.name)
	}

	if tag.attributes == nil {
		t.Error("Expected attributes map to be initialized")
	}

	if tag.classes == nil {
		t.Error("Expected classes slice to be initialized")
	}

	if tag.styles == nil {
		t.Error("Expected styles slice to be initialized")
	}
}

func TestAddStyle(t *testing.T) {
	tag := NewHTMLTag("div")

	result := tag.AddStyle("color", "red")

	// Should return the same instance for chaining
	if result != tag {
		t.Error("AddStyle should return the same instance for chaining")
	}

	if len(tag.styles) != 1 {
		t.Errorf("Expected 1 style, got %d", len(tag.styles))
	}

	style := tag.styles[0]
	if style.Name != "color" {
		t.Errorf("Expected style name 'color', got '%s'", style.Name)
	}

	if style.Value != "red" {
		t.Errorf("Expected style value 'red', got '%s'", style.Value)
	}
}

func TestMaybeAddStyle(t *testing.T) {
	tag := NewHTMLTag("div")

	// Test with nil value - should not add style
	result := tag.MaybeAddStyle("color", nil)
	if result != tag {
		t.Error("MaybeAddStyle should return the same instance for chaining")
	}
	if len(tag.styles) != 0 {
		t.Error("MaybeAddStyle with nil should not add style")
	}

	// Test with empty value - should not add style
	empty := ""
	tag.MaybeAddStyle("background", &empty)
	if len(tag.styles) != 0 {
		t.Error("MaybeAddStyle with empty string should not add style")
	}

	// Test with valid value - should add style
	value := "blue"
	tag.MaybeAddStyle("color", &value)
	if len(tag.styles) != 1 {
		t.Error("MaybeAddStyle with valid value should add style")
	}

	style := tag.styles[0]
	if style.Name != "color" || style.Value != "blue" {
		t.Errorf("Expected style 'color: blue', got '%s: %s'", style.Name, style.Value)
	}
}

func TestAddAttribute(t *testing.T) {
	tag := NewHTMLTag("div")

	result := tag.AddAttribute("class", "container")

	// Should return the same instance for chaining
	if result != tag {
		t.Error("AddAttribute should return the same instance for chaining")
	}

	if len(tag.attributes) != 1 {
		t.Errorf("Expected 1 attribute, got %d", len(tag.attributes))
	}

	attr := tag.attributes[0]
	if attr.Name != "class" {
		t.Errorf("Expected attribute name 'class', got '%s'", attr.Name)
	}
	if attr.Value != "container" {
		t.Errorf("Expected attribute value 'container', got '%s'", attr.Value)
	}
}

func TestMaybeAddAttribute(t *testing.T) {
	tag := NewHTMLTag("div")

	// Test with nil value - should not add attribute
	result := tag.MaybeAddAttribute("id", nil)
	if result != tag {
		t.Error("MaybeAddAttribute should return the same instance for chaining")
	}
	if len(tag.attributes) != 0 {
		t.Error("MaybeAddAttribute with nil should not add attribute")
	}

	// Test with empty value - should not add attribute
	empty := ""
	tag.MaybeAddAttribute("class", &empty)
	if len(tag.attributes) != 0 {
		t.Error("MaybeAddAttribute with empty string should not add attribute")
	}

	// Test with valid value - should add attribute
	value := "test-id"
	tag.MaybeAddAttribute("id", &value)
	if len(tag.attributes) != 1 {
		t.Error("MaybeAddAttribute with valid value should add attribute")
	}

	attr := tag.attributes[0]
	if attr.Name != "id" {
		t.Errorf("Expected attribute name 'id', got '%s'", attr.Name)
	}
	if attr.Value != "test-id" {
		t.Errorf("Expected attribute value 'test-id', got '%s'", attr.Value)
	}
}

func TestAddClass(t *testing.T) {
	tag := NewHTMLTag("div")

	result := tag.AddClass("container")

	// Should return the same instance for chaining
	if result != tag {
		t.Error("AddClass should return the same instance for chaining")
	}

	if len(tag.classes) != 1 {
		t.Errorf("Expected 1 class, got %d", len(tag.classes))
	}

	if tag.classes[0] != "container" {
		t.Errorf("Expected class 'container', got '%s'", tag.classes[0])
	}
}

func TestRenderOpen(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*HTMLTag) *HTMLTag
		expected string
	}{
		{
			name:     "simple tag",
			setup:    func(tag *HTMLTag) *HTMLTag { return tag },
			expected: "<div>",
		},
		{
			name: "tag with attribute",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.AddAttribute("id", "test")
			},
			expected: `<div id="test">`,
		},
		{
			name: "tag with class",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.AddClass("container")
			},
			expected: `<div class="container">`,
		},
		{
			name: "tag with multiple classes",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.AddClass("container").AddClass("responsive")
			},
			expected: `<div class="container responsive">`,
		},
		{
			name: "tag with style",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.AddStyle("color", "red")
			},
			expected: `<div style="color:red;">`,
		},
		{
			name: "tag with multiple styles",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.AddStyle("color", "red").AddStyle("background", "blue")
			},
			expected: `<div style="color:red;background:blue;">`,
		},
		{
			name: "tag with everything",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.
					AddAttribute("id", "test").
					AddClass("container").
					AddClass("responsive").
					AddStyle("color", "red").
					AddStyle("margin", "10px")
			},
			expected: `<div id="test" class="container responsive" style="color:red;margin:10px;">`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := NewHTMLTag("div")
			tag = tt.setup(tag)

			var buf strings.Builder
			err := tag.RenderOpen(&buf)
			if err != nil {
				t.Fatalf("RenderOpen failed: %v", err)
			}
			result := buf.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestRenderClose(t *testing.T) {
	tag := NewHTMLTag("div")
	var buf strings.Builder
	err := tag.RenderClose(&buf)
	if err != nil {
		t.Fatalf("RenderClose failed: %v", err)
	}
	result := buf.String()
	expected := "</div>"

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestRenderSelfClosing(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*HTMLTag) *HTMLTag
		expected string
	}{
		{
			name:     "simple self-closing tag",
			setup:    func(tag *HTMLTag) *HTMLTag { return tag },
			expected: "<img />",
		},
		{
			name: "self-closing tag with attributes",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.AddAttribute("src", "image.jpg").AddAttribute("alt", "test")
			},
			expected: `<img src="image.jpg" alt="test" />`,
		},
		{
			name: "self-closing tag with styles",
			setup: func(tag *HTMLTag) *HTMLTag {
				return tag.AddStyle("width", "100px").AddStyle("height", "100px")
			},
			expected: `<img style="width:100px;height:100px;" />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := NewHTMLTag("img")
			tag = tt.setup(tag)

			var buf strings.Builder
			err := tag.RenderSelfClosing(&buf)
			if err != nil {
				t.Fatalf("RenderSelfClosing failed: %v", err)
			}
			result := buf.String()
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestChaining(t *testing.T) {
	tag := NewHTMLTag("div").
		AddAttribute("id", "test").
		AddAttribute("data-value", "123").
		AddClass("container").
		AddClass("responsive").
		AddStyle("color", "red").
		AddStyle("background", "blue").
		AddStyle("margin", "10px")

	// Verify all values were set
	if len(tag.attributes) != 2 {
		t.Errorf("Expected 2 attributes, got %d", len(tag.attributes))
	}

	if len(tag.classes) != 2 {
		t.Errorf("Expected 2 classes, got %d", len(tag.classes))
	}

	if len(tag.styles) != 3 {
		t.Errorf("Expected 3 styles, got %d", len(tag.styles))
	}

	// Test rendering
	var buf strings.Builder
	err := tag.RenderOpen(&buf)
	if err != nil {
		t.Fatalf("RenderOpen failed: %v", err)
	}
	result := buf.String()
	expectedParts := []string{
		`id="test"`,
		`data-value="123"`,
		`class="container responsive"`,
		`style="color:red;background:blue;margin:10px;"`,
	}

	for _, part := range expectedParts {
		if !strings.Contains(result, part) {
			t.Errorf("Expected result to contain '%s', got '%s'", part, result)
		}
	}
}

func TestStylePropertyStruct(t *testing.T) {
	prop := StyleProperty{
		Name:  "color",
		Value: "red",
	}

	if prop.Name != "color" {
		t.Errorf("Expected Name 'color', got '%s'", prop.Name)
	}

	if prop.Value != "red" {
		t.Errorf("Expected Value 'red', got '%s'", prop.Value)
	}
}

// TestStyleOrdering verifies that styles maintain insertion order
func TestStyleOrdering(t *testing.T) {
	tag := NewHTMLTag("div").
		AddStyle("z-index", "1").
		AddStyle("color", "red").
		AddStyle("background", "blue").
		AddStyle("margin", "10px")

	var buf strings.Builder
	err := tag.RenderOpen(&buf)
	if err != nil {
		t.Fatalf("RenderOpen failed: %v", err)
	}
	result := buf.String()
	expected := `<div style="z-index:1;color:red;background:blue;margin:10px;">`

	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// TestEscapeAttributeValue verifies that attribute values are escaped on output
// for the characters that corrupt a double-quoted attribute (" < >), while
// ampersands - including values that already carry valid &...; entities, such as
// the pre-escaped mj-section background URL - are left untouched so they are not
// double-encoded.
func TestEscapeAttributeValue(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "plain value is unchanged", value: "hello world", want: "hello world"},
		{name: "double quote is encoded", value: `27" monitor`, want: `27&quot; monitor`},
		{name: "angle brackets are encoded", value: `tag <b> here`, want: `tag &lt;b&gt; here`},
		{
			name:  "bare ampersand is left as-is",
			value: `a=1&b=2`,
			want:  `a=1&b=2`,
		},
		{
			name:  "existing entity is not double-escaped",
			value: `https://x.test/i.png?w=500&amp;auto=format`,
			want:  `https://x.test/i.png?w=500&amp;auto=format`,
		},
		{
			name:  "mixed quote angle and ampersand",
			value: `he said "a<b" & "c>d"`,
			want:  `he said &quot;a&lt;b&quot; & &quot;c&gt;d&quot;`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeAttributeValue(tt.value); got != tt.want {
				t.Errorf("escapeAttributeValue(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

// TestRenderSelfClosingEscapesAttributeValue verifies the escaping is wired into
// the rendering path, so an alt text containing a double quote does not truncate
// the attribute.
func TestRenderSelfClosingEscapesAttributeValue(t *testing.T) {
	var buf strings.Builder
	err := NewHTMLTag("img").
		AddAttribute("src", "image.jpg").
		AddAttribute("alt", `27" monitor`).
		RenderSelfClosing(&buf)
	if err != nil {
		t.Fatalf("RenderSelfClosing failed: %v", err)
	}

	result := buf.String()
	expected := `<img src="image.jpg" alt="27&quot; monitor" />`
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

// failingStringWriter fails the write of one specific string, so a test can
// drive the error return that follows that write.
type failingStringWriter struct {
	strings.Builder
	failOn string
}

var errWriteFailed = errors.New("write failed")

func (w *failingStringWriter) WriteString(s string) (int, error) {
	if s == w.failOn {
		return 0, errWriteFailed
	}
	return w.Builder.WriteString(s)
}

func TestRenderOpenReturnsAttributeValueWriteError(t *testing.T) {
	tag := NewHTMLTag("img").AddAttribute("alt", `27" monitor`)

	// The writer only fails on the escaped form of the value, so getting the error
	// back also shows that the value reaches the writer escaped.
	err := tag.RenderOpen(&failingStringWriter{failOn: "27&quot; monitor"})
	if !errors.Is(err, errWriteFailed) {
		t.Fatalf("RenderOpen() error = %v, want %v", err, errWriteFailed)
	}
}
