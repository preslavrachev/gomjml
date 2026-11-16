package mjml

import (
	"errors"
	"strings"
	"testing"
)

// TestOutputLargerThan16MBIsRefused checks that a template whose attribute
// value is written onto every element cannot make one render exceed 16 MB.
func TestOutputLargerThan16MBIsRefused(t *testing.T) {
	//GIVEN: a 100 KB mj-all font-family value that every one of 2,400 mj-text elements receives.
	value := strings.Repeat("z", 100_000)
	texts := strings.Repeat(`<mj-text>t</mj-text>`, 2400)
	template := `<mjml><mj-head><mj-attributes><mj-all font-family="` + value + `"/></mj-attributes></mj-head>` +
		`<mj-body><mj-section><mj-column>` + texts + `</mj-column></mj-section></mj-body></mjml>`

	//WHEN: the template is rendered.
	_, err := Render(template)

	//THEN: the render is refused with ErrOutputTooLarge.
	if !errors.Is(err, ErrOutputTooLarge) {
		t.Fatalf("got error %v, want ErrOutputTooLarge", err)
	}
}
