package mjml

import (
	"strings"
	"testing"
)

// The widths are MJML 4.16.1's for the same wrappers.
func TestWrapperPaddingNarrowsSections(t *testing.T) {
	wrappers := []struct{ body, attrs, want string }{
		{"", `padding="5%"`, "590px"},
		{"", `padding="1px 2px 3px"`, "596px"},
		{"", `padding-left="50%"`, "550px"},
		{"", `padding="0 20px" padding-right="10px"`, "570px"},
		{"", `border-left="3px solid red" padding="0 10px"`, "577px"},
		{"", `full-width="full-width" padding="0 30px"`, "540px"},
		{"", `full-width="full-width" padding="0 30px" padding-left="5px"`, "565px"},
		{` width="400px"`, `padding="0 20px"`, "360px"},
	}
	for _, w := range wrappers {
		html, err := Render(`<mjml><mj-body` + w.body + `><mj-wrapper ` + w.attrs + `><mj-section><mj-column><mj-text>x</mj-text></mj-column></mj-section></mj-wrapper></mj-body></mjml>`)
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range []string{"max-width:" + w.want, "vertical-align:top;width:" + w.want} {
			if !strings.Contains(html, want) {
				t.Errorf("<mj-wrapper %s>: the section lacks %s", w.attrs, want)
			}
		}
	}
}
