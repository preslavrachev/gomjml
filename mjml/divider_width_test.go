package mjml

import (
	"regexp"
	"testing"
)

// The widths are MJML 4.16.1's for the same dividers.
func TestDividerOutlookWidth(t *testing.T) {
	dividers := []struct{ attrs, want string }{
		{`padding="5%"`, "590px"},
		{`padding-left="50%"`, "525px"},
		{`padding="0 10px" padding-right="0px" width="50%"`, "295px"},
		{`width="33%" padding="0 7px"`, "193.38px"},
		// MJML multiplies by 0.33, not by 33 and then divides.
		{`width="33%" padding="10px 5%"`, "194.70000000000002px"},
		{`width="120px"`, "120px"},
		{`padding="1px 2px 3px"`, "596px"},
	}
	for _, d := range dividers {
		html, err := Render(`<mjml><mj-body><mj-section><mj-column><mj-divider ` + d.attrs + ` /></mj-column></mj-section></mj-body></mjml>`)
		if err != nil {
			t.Fatal(err)
		}
		got := regexp.MustCompile(`<table align="[^"]*" border="0" cellpadding="0" cellspacing="0" style="border-top:[^"]*;width:([^;]*);" role="presentation" width="([^"]*)"`).FindStringSubmatch(html)
		if got == nil {
			t.Fatalf("<mj-divider %s>: no Outlook table in\n%s", d.attrs, html)
		}
		if got[1] != d.want || got[2] != d.want {
			t.Errorf("<mj-divider %s>: Outlook table width %s, width=%q; want %s", d.attrs, got[1], got[2], d.want)
		}
	}
}
