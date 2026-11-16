package mjml

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/preslavrachev/gomjml/parser"
)

// TestInlineStyleBudgetStopsAmplifiedRules checks that a small template cannot
// make one render write hundreds of megabytes by repeating a large inline
// mj-style rule on many elements.
func TestInlineStyleBudgetStopsAmplifiedRules(t *testing.T) {
	//GIVEN: a 196 KB template whose class carries 16,000 declarations and is applied to 4,000 elements.
	template := amplifiedInlineStyleTemplate(16000, 4000)

	//WHEN: the template is rendered.
	_, err := Render(template)

	//THEN: the render is refused with ErrInlineStyleBudgetExceeded.
	if !errors.Is(err, ErrInlineStyleBudgetExceeded) {
		t.Fatalf("got error %v, want ErrInlineStyleBudgetExceeded", err)
	}
}

// TestInlineStyleBudgetKeepsNormalTemplates checks that an ordinary template
// with an inline mj-style block still renders, so the limit does not touch
// real emails.
func TestInlineStyleBudgetKeepsNormalTemplates(t *testing.T) {
	//GIVEN: the mj-style golden template, which uses an inline mj-style block.
	template, err := os.ReadFile("testdata/mj-style.mjml")
	if err != nil {
		t.Fatal(err)
	}

	//WHEN: the template is rendered.
	_, err = Render(string(template))

	//THEN: the render succeeds.
	if err != nil {
		t.Fatalf("got error %v, want success", err)
	}
}

// amplifiedInlineStyleTemplate returns a template whose single inline mj-style
// class has the given number of declarations, applied to the given number of
// paragraphs.
func amplifiedInlineStyleTemplate(declarations, elements int) string {
	var css strings.Builder
	for i := range declarations {
		fmt.Fprintf(&css, "p%d:1;", i)
	}
	html := strings.Repeat(`<p class="a"></p>`, elements)
	return `<mjml><mj-head><mj-style inline="inline">.a{` + css.String() + `}</mj-style></mj-head>` +
		`<mj-body><mj-section><mj-column><mj-text>` + html + `</mj-text></mj-column></mj-section></mj-body></mjml>`
}

// TestCreateComponentWithoutOptionsKeepsTheInlineBudget checks that a caller who
// builds a component without render options still gets the inline style limit.
func TestCreateComponentWithoutOptionsKeepsTheInlineBudget(t *testing.T) {
	//GIVEN: the parsed 16,000 x 4,000 amplified template and no render options.
	ast, err := parser.ParseMJML(amplifiedInlineStyleTemplate(16000, 4000))
	if err != nil {
		t.Fatal(err)
	}

	//WHEN: the component is created with nil options and rendered into a builder.
	component, err := CreateComponent(ast, nil)
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	err = component.Render(&out)

	//THEN: the render is refused with ErrInlineStyleBudgetExceeded.
	if !errors.Is(err, ErrInlineStyleBudgetExceeded) {
		t.Fatalf("got error %v, want ErrInlineStyleBudgetExceeded", err)
	}
}
