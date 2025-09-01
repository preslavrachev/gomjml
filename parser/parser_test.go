package parser

import (
	"strings"
	"testing"
)

func TestParseMJML(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantTag string
		wantErr bool
	}{
		{
			name:    "basic mjml",
			input:   `<mjml><mj-body><mj-text>Hello</mj-text></mj-body></mjml>`,
			wantTag: "mjml",
			wantErr: false,
		},
		{
			name:    "with attributes",
			input:   `<mjml version="4.0"><mj-head><mj-title>Test</mj-title></mj-head></mjml>`,
			wantTag: "mjml",
			wantErr: false,
		},
		{
			name:    "empty input",
			input:   "",
			wantTag: "",
			wantErr: true,
		},
		{
			name:    "invalid xml",
			input:   `<mjml><mj-body><mj-text>Hello</mj-body></mjml>`,
			wantTag: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseMJML(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseMJML() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseMJML() error = %v", err)
				return
			}

			if node.GetTagName() != tt.wantTag {
				t.Errorf("ParseMJML() tag = %v, want %v", node.GetTagName(), tt.wantTag)
			}
		})
	}
}

func TestMJMLNode_GetAttribute(t *testing.T) {
	input := `<mjml version="4.0" lang="en"><mj-head></mj-head></mjml>`
	node, err := ParseMJML(input)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}

	tests := []struct {
		name string
		attr string
		want string
	}{
		{"existing attribute", "version", "4.0"},
		{"another existing attribute", "lang", "en"},
		{"non-existing attribute", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := node.GetAttribute(tt.attr)
			if got != tt.want {
				t.Errorf("GetAttribute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMJMLNode_FindFirstChild(t *testing.T) {
	input := `<mjml><mj-head><mj-title>Test</mj-title></mj-head><mj-body><mj-text>Hello</mj-text></mj-body></mjml>`
	node, err := ParseMJML(input)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}

	tests := []struct {
		name    string
		tagName string
		wantTag string
		wantNil bool
	}{
		{"find head", "mj-head", "mj-head", false},
		{"find body", "mj-body", "mj-body", false},
		{"find non-existing", "mj-section", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := node.FindFirstChild(tt.tagName)

			if tt.wantNil {
				if child != nil {
					t.Errorf("FindFirstChild() expected nil but got %v", child.GetTagName())
				}
				return
			}

			if child == nil {
				t.Errorf("FindFirstChild() expected child but got nil")
				return
			}

			if child.GetTagName() != tt.wantTag {
				t.Errorf("FindFirstChild() tag = %v, want %v", child.GetTagName(), tt.wantTag)
			}
		})
	}
}

func TestMJMLNode_GetTextContent(t *testing.T) {
	input := `<mjml><mj-body><mj-text>  Hello World  </mj-text></mj-body></mjml>`
	node, err := ParseMJML(input)
	if err != nil {
		t.Fatalf("ParseMJML() error = %v", err)
	}

	textNode := node.FindFirstChild("mj-body").FindFirstChild("mj-text")
	if textNode == nil {
		t.Fatal("Could not find mj-text node")
	}

	got := textNode.GetTextContent()
	want := "Hello World"

	if got != want {
		t.Errorf("GetTextContent() = %q, want %q", got, want)
	}
}

func TestMJMLRaw_SingleVoidElement(t *testing.T) {
	mjml := `<mjml>
<mj-body>
	<mj-raw>
		<p>Text before void element</p>
		<img src="test.jpg" alt="test">
		<p>Text after void element</p>
	</mj-raw>
</mj-body>
</mjml>`

	node, err := ParseMJML(mjml)
	if err != nil {
		t.Fatalf("MJML with single void element should parse without error: %v", err)
	}

	rawElement := node.FindFirstChild("mj-body").FindFirstChild("mj-raw")
	if rawElement == nil {
		t.Fatal("Should find mj-raw element")
	}

	if !strings.Contains(rawElement.Text, "<img") {
		t.Error("Raw content should preserve img tag")
	}
}

func TestMJMLRaw_MultipleVoidElements(t *testing.T) {
	mjml := `<mjml>
<mj-body>
	<mj-raw>
		<p>Paragraph with <br> line break</p>
		<hr>
		<img src="test.jpg" alt="test">
		<input type="text" name="test">
		<p>End paragraph</p>
	</mj-raw>
</mj-body>
</mjml>`

	node, err := ParseMJML(mjml)
	if err != nil {
		t.Fatalf("MJML with multiple void elements should parse without error: %v", err)
	}

	rawElement := node.FindFirstChild("mj-body").FindFirstChild("mj-raw")
	if rawElement == nil {
		t.Fatal("Should find mj-raw element")
	}

	voidTags := []string{"<br", "<hr", "<img", "<input"}
	for _, tag := range voidTags {
		if !strings.Contains(rawElement.Text, tag) {
			t.Errorf("Raw content should preserve %s tag", tag)
		}
	}
}

func TestMJMLRaw_NestedVoidElements(t *testing.T) {
	mjml := `<mjml>
<mj-body>
	<mj-raw>
		<div>
			<p>Text <img src="icon.png" alt="icon"> with image</p>
			<p>Another <br> paragraph</p>
		</div>
	</mj-raw>
</mj-body>
</mjml>`

	node, err := ParseMJML(mjml)
	if err != nil {
		t.Fatalf("MJML with nested void elements should parse without error: %v", err)
	}

	rawElement := node.FindFirstChild("mj-body").FindFirstChild("mj-raw")
	if rawElement == nil {
		t.Fatal("Should find mj-raw element")
	}

	if !strings.Contains(rawElement.Text, "<div>") || !strings.Contains(rawElement.Text, "</div>") {
		t.Error("Raw content should preserve div nesting")
	}

	if !strings.Contains(rawElement.Text, "<img") {
		t.Error("Raw content should preserve img tag")
	}

	if !strings.Contains(rawElement.Text, "<br") {
		t.Error("Raw content should preserve br tag")
	}
}

func TestMJMLRaw_PreservesCompleteHTMLStructure(t *testing.T) {
	mjml := `<mjml>
<mj-body>
	<mj-raw>
		<div class="container">
			<img src="test.jpg" alt="test" />
			<p>Content after void element</p>
			<hr />
			<p>Final paragraph</p>
		</div>
	</mj-raw>
</mj-body>
</mjml>`

	node, err := ParseMJML(mjml)
	if err != nil {
		t.Fatalf("Parse should succeed: %v", err)
	}

	rawElement := node.FindFirstChild("mj-body").FindFirstChild("mj-raw")
	if rawElement == nil {
		t.Fatal("Should find mj-raw element")
	}

	content := rawElement.Text

	// The parser must preserve the complete HTML structure including all content after void elements
	requiredElements := []string{
		`<div class="container">`,
		`</div>`,
		`<img src="test.jpg" alt="test"`,
		`<p>Content after void element</p>`,
		`<hr`,
		`<p>Final paragraph</p>`,
	}

	t.Logf("Raw content: %q", content)

	for _, required := range requiredElements {
		if !strings.Contains(content, required) {
			t.Errorf("Raw content missing required element: %s\nActual content: %s", required, content)
		}
	}
}

func TestEscapeAttributeAmpersands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "URL with query parameters",
			input:    `<mj-image src="https://example.com?param1=value1&param2=value2" />`,
			expected: `<mj-image src="https://example.com?param1=value1&amp;param2=value2" />`,
		},
		{
			name:     "Multiple URL parameters",
			input:    `<mj-button href="https://example.com?a=1&b=2&c=3&d=4" />`,
			expected: `<mj-button href="https://example.com?a=1&amp;b=2&amp;c=3&amp;d=4" />`,
		},
		{
			name:     "Mixed valid entities and raw ampersands",
			input:    `<mj-text>Hello &amp; World</mj-text><mj-image src="https://example.com?test=1&param=2" />`,
			expected: `<mj-text>Hello &amp; World</mj-text><mj-image src="https://example.com?test=1&amp;param=2" />`,
		},
		{
			name:     "Single quotes in attributes",
			input:    `<mj-image src='https://example.com?param1=value1&param2=value2' />`,
			expected: `<mj-image src='https://example.com?param1=value1&amp;param2=value2' />`,
		},
		{
			name:     "Already escaped ampersands should not be double-escaped",
			input:    `<mj-image src="https://example.com?param1=value1&amp;param2=value2" />`,
			expected: `<mj-image src="https://example.com?param1=value1&amp;param2=value2" />`,
		},
		{
			name:     "Multiple attributes with ampersands",
			input:    `<mj-button href="https://example.com?a=1&b=2" title="Click & Win" />`,
			expected: `<mj-button href="https://example.com?a=1&amp;b=2" title="Click &amp; Win" />`,
		},
		{
			name:     "Valid HTML entities should not be escaped",
			input:    `<mj-text content="&lt;b&gt;Bold&lt;/b&gt; &amp; &copy; 2024" />`,
			expected: `<mj-text content="&lt;b&gt;Bold&lt;/b&gt; &amp; &copy; 2024" />`,
		},
		{
			name:     "Numeric character references",
			input:    `<mj-text content="Price: &#36;100 &amp; tax &#160;included" />`,
			expected: `<mj-text content="Price: &#36;100 &amp; tax &#160;included" />`,
		},
		{
			name:     "Hexadecimal character references",
			input:    `<mj-text content="Unicode: &#x20AC; &amp; more" />`,
			expected: `<mj-text content="Unicode: &#x20AC; &amp; more" />`,
		},
		{
			name:     "Complex URL with fragment and special characters",
			input:    `<mj-button href="https://example.com/path?query=test&filter=active#section&more" />`,
			expected: `<mj-button href="https://example.com/path?query=test&amp;filter=active#section&amp;more" />`,
		},
		{
			name:     "Multiple components on same line",
			input:    `<mj-image src="https://img.com?w=300&h=200" /><mj-button href="https://btn.com?action=click&ref=email" />`,
			expected: `<mj-image src="https://img.com?w=300&amp;h=200" /><mj-button href="https://btn.com?action=click&amp;ref=email" />`,
		},
		{
			name:     "Attributes without quotes should be ignored",
			input:    `<mj-image src=https://example.com?param1=value1&param2=value2 width="300px" />`,
			expected: `<mj-image src=https://example.com?param1=value1&param2=value2 width="300px" />`,
		},
		{
			name:     "Empty attributes",
			input:    `<mj-image src="" alt="" />`,
			expected: `<mj-image src="" alt="" />`,
		},
		{
			name:     "Attributes with spaces around equals",
			input:    `<mj-image src = "https://example.com?a=1&b=2" width= "300px" />`,
			expected: `<mj-image src = "https://example.com?a=1&amp;b=2" width= "300px" />`,
		},
		{
			name:     "No attributes to process",
			input:    `<mj-text>Just some content with & ampersand</mj-text>`,
			expected: `<mj-text>Just some content with & ampersand</mj-text>`,
		},
		{
			name:     "Social media URLs",
			input:    `<mj-social-element href="https://www.facebook.com/sharer/sharer.php?u=https://example.com&t=My Title" />`,
			expected: `<mj-social-element href="https://www.facebook.com/sharer/sharer.php?u=https://example.com&amp;t=My Title" />`,
		},
		{
			name:     "Carousel images with thumbnails",
			input:    `<mj-carousel-image src="https://img.com/main.jpg" thumbnails-src="https://img.com/thumb.jpg?w=100&h=100" />`,
			expected: `<mj-carousel-image src="https://img.com/main.jpg" thumbnails-src="https://img.com/thumb.jpg?w=100&amp;h=100" />`,
		},
		{
			name:     "Navbar with base URL",
			input:    `<mj-navbar base-url="https://example.com" href="/path?param=value&other=test" />`,
			expected: `<mj-navbar base-url="https://example.com" href="/path?param=value&amp;other=test" />`,
		},
		{
			name:     "Hero background URL",
			input:    `<mj-hero background-url="https://images.com/hero.jpg?quality=high&format=webp" />`,
			expected: `<mj-hero background-url="https://images.com/hero.jpg?quality=high&amp;format=webp" />`,
		},
		{
			name:     "Accordion icon URLs",
			input:    `<mj-accordion icon-wrapped-url="https://icons.com/plus.svg?color=blue&size=16" icon-unwrapped-url="https://icons.com/minus.svg?color=red&size=16" />`,
			expected: `<mj-accordion icon-wrapped-url="https://icons.com/plus.svg?color=blue&amp;size=16" icon-unwrapped-url="https://icons.com/minus.svg?color=red&amp;size=16" />`,
		},
		{
			name:     "The original failing case",
			input:    `<mj-image width="300px" src="https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&text=LOGO" />`,
			expected: `<mj-image width="300px" src="https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&amp;text=LOGO" />`,
		},
		{
			name:     "Edge case: Attribute value contains quotes",
			input:    `<mj-text title="Say \"Hello\" & welcome" />`,
			expected: `<mj-text title="Say \"Hello\" &amp; welcome" />`,
		},
		{
			name:     "Edge case: Mixed quote types",
			input:    `<mj-button href='https://example.com?msg=Hello&name=World' title="Click & Go" />`,
			expected: `<mj-button href='https://example.com?msg=Hello&amp;name=World' title="Click &amp; Go" />`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeAttributeAmpersands(tt.input)
			if result != tt.expected {
				t.Errorf("escapeAttributeAmpersands() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestEscapeAmperands(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple URL with parameters",
			input:    "https://example.com?param1=value1&param2=value2",
			expected: "https://example.com?param1=value1&amp;param2=value2",
		},
		{
			name:     "Already escaped ampersands",
			input:    "https://example.com?param1=value1&amp;param2=value2",
			expected: "https://example.com?param1=value1&amp;param2=value2",
		},
		{
			name:     "Mixed escaped and unescaped",
			input:    "https://example.com?a=1&b=2&amp;c=3&d=4",
			expected: "https://example.com?a=1&amp;b=2&amp;c=3&amp;d=4",
		},
		{
			name:     "Valid HTML entities",
			input:    "&lt;b&gt;Bold&lt;/b&gt; &amp; &copy; 2024",
			expected: "&lt;b&gt;Bold&lt;/b&gt; &amp; &copy; 2024",
		},
		{
			name:     "Numeric character references",
			input:    "Price: &#36;100 & &#160;tax",
			expected: "Price: &#36;100 &amp; &#160;tax",
		},
		{
			name:     "Hexadecimal character references",
			input:    "Unicode: &#x20AC; & more &#xA0;space",
			expected: "Unicode: &#x20AC; &amp; more &#xA0;space",
		},
		{
			name:     "Raw ampersands only",
			input:    "Tom & Jerry & Friends",
			expected: "Tom &amp; Jerry &amp; Friends",
		},
		{
			name:     "Complex mix",
			input:    "Visit &lt;a href=\"https://example.com?ref=email&utm_source=newsletter\"&gt;our site&lt;/a&gt; &amp; enjoy!",
			expected: "Visit &lt;a href=\"https://example.com?ref=email&amp;utm_source=newsletter\"&gt;our site&lt;/a&gt; &amp; enjoy!",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only ampersand",
			input:    "&",
			expected: "&amp;",
		},
		{
			name:     "Incomplete entity",
			input:    "Test &incomplete entity &amp; valid",
			expected: "Test &amp;incomplete entity &amp; valid",
		},
		{
			name:     "Entity-like but invalid",
			input:    "&fake; &notreal; &amp; &copy;",
			expected: "&amp;fake; &amp;notreal; &amp; &copy;",
		},
		{
			name:     "Case sensitivity",
			input:    "&AMP; &Amp; &amp; &COPY; &copy;",
			expected: "&amp;AMP; &amp;Amp; &amp; &amp;COPY; &copy;",
		},
		{
			name:     "Multiple consecutive ampersands",
			input:    "Test && multiple &&&& ampersands",
			expected: "Test &amp;&amp; multiple &amp;&amp;&amp;&amp; ampersands",
		},
		{
			name:     "URL fragment with ampersands",
			input:    "https://example.com/page#section&subsection&details",
			expected: "https://example.com/page#section&amp;subsection&amp;details",
		},
		{
			name:     "The exact failing string from test",
			input:    "https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&text=LOGO",
			expected: "https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&amp;text=LOGO",
		},
		{
			name:     "Common Google Fonts URL pattern",
			input:    "https://fonts.googleapis.com/css?family=Roboto:400,700&display=swap",
			expected: "https://fonts.googleapis.com/css?family=Roboto:400,700&amp;display=swap",
		},
		{
			name:     "Analytics URL with multiple parameters",
			input:    "https://analytics.com?utm_source=email&utm_medium=newsletter&utm_campaign=summer&utm_content=button",
			expected: "https://analytics.com?utm_source=email&amp;utm_medium=newsletter&amp;utm_campaign=summer&amp;utm_content=button",
		},
		{
			name:     "Edge case: Ampersand at start",
			input:    "&start=here&end=there",
			expected: "&amp;start=here&amp;end=there",
		},
		{
			name:     "Edge case: Ampersand at end",
			input:    "https://example.com?param=value&",
			expected: "https://example.com?param=value&amp;",
		},
		{
			name:     "All valid common entities",
			input:    "&amp; &lt; &gt; &quot; &apos; &nbsp; &copy; &reg; &trade; &ndash; &mdash; &hellip;",
			expected: "&amp; &lt; &gt; &quot; &apos; &nbsp; &copy; &reg; &trade; &ndash; &mdash; &hellip;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := escapeAmperands(tt.input)
			if result != tt.expected {
				t.Errorf("escapeAmperands() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPreprocessHTMLEntitiesIntegration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "Full MJML with URL parameters",
			input: `<mjml>
  <mj-body>
    <mj-section>
      <mj-column>
        <mj-image width="300px" src="https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&text=LOGO" />
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>`,
			expected: `<mjml>
  <mj-body>
    <mj-section>
      <mj-column>
        <mj-image width="300px" src="https://placehold.co/150x60/E3F2FD/1976D2?font=playfair-display&amp;text=LOGO" />
      </mj-column>
    </mj-section>
  </mj-body>
</mjml>`,
		},
		{
			name: "Complex MJML with multiple components and entities",
			input: `<mjml>
  <mj-body>
    <mj-text>&copy; 2024 Company &amp; Co.</mj-text>
    <mj-button href="https://example.com?utm_source=email&utm_medium=newsletter">Visit &amp; Buy</mj-button>
    <mj-image src="https://images.com/logo.png?w=300&h=150&quality=high" alt="Logo &amp; Tagline" />
  </mj-body>
</mjml>`,
			expected: `<mjml>
  <mj-body>
    <mj-text>© 2024 Company &amp; Co.</mj-text>
    <mj-button href="https://example.com?utm_source=email&amp;utm_medium=newsletter">Visit &amp; Buy</mj-button>
    <mj-image src="https://images.com/logo.png?w=300&amp;h=150&amp;quality=high" alt="Logo &amp; Tagline" />
  </mj-body>
</mjml>`,
		},
		{
			name: "Social component URLs",
			input: `<mjml>
  <mj-body>
    <mj-social>
      <mj-social-element name="facebook" href="https://facebook.com/share?u=https://example.com&t=Check this out" src="https://icons.com/fb.svg?size=32&color=blue" />
      <mj-social-element name="twitter" href="https://twitter.com/intent/tweet?text=Hello&url=https://example.com" />
    </mj-social>
  </mj-body>
</mjml>`,
			expected: `<mjml>
  <mj-body>
    <mj-social>
      <mj-social-element name="facebook" href="https://facebook.com/share?u=https://example.com&amp;t=Check this out" src="https://icons.com/fb.svg?size=32&amp;color=blue" />
      <mj-social-element name="twitter" href="https://twitter.com/intent/tweet?text=Hello&amp;url=https://example.com" />
    </mj-social>
  </mj-body>
</mjml>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := preprocessHTMLEntities(tt.input)
			if result != tt.expected {
				t.Errorf("preprocessHTMLEntities() = %v, want %v", result, tt.expected)
			}
		})
	}
}
