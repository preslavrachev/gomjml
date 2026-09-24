package mjml

import (
	"regexp"
	"slices"
	"testing"
)

var mediaQuery = regexp.MustCompile(`\((?:min|max)-width:[^)]*\)`)

// The expected queries are what MJML 4.16.1 renders for the same input: the
// column query and its Thunderbird twin at the breakpoint, and the image and
// navbar mobile queries one pixel below it.
func TestBreakpointSetsEveryMediaQuery(t *testing.T) {
	body := `<mj-body><mj-section><mj-column>
		<mj-image src="https://example.com/a.png" fluid-on-mobile="true" />
		<mj-navbar hamburger="hamburger"><mj-navbar-link href="/a">A</mj-navbar-link></mj-navbar>
	</mj-column></mj-section></mj-body>`

	tests := []struct {
		name string
		head string
		want []string
	}{
		{
			name: "default",
			want: []string{"(min-width:480px)", "(min-width:480px)", "(max-width:479px)", "(max-width:479px)"},
		},
		{
			name: "last breakpoint wins",
			head: `<mj-breakpoint width="400px" /><mj-breakpoint width="600px" />`,
			want: []string{"(min-width:600px)", "(min-width:600px)", "(max-width:599px)", "(max-width:599px)"},
		},
		{
			// MJML warns about the unit but still uses the value, and takes
			// the lower breakpoint from its leading number.
			name: "non-px width",
			head: `<mj-breakpoint width="37em" />`,
			want: []string{"(min-width:37em)", "(min-width:37em)", "(max-width:36px)", "(max-width:36px)"},
		},
		{
			name: "leading whitespace",
			head: `<mj-breakpoint width=" 600px" />`,
			want: []string{"(min-width: 600px)", "(min-width: 600px)", "(max-width:599px)", "(max-width:599px)"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			html, err := Render(`<mjml><mj-head>` + tc.head + `</mj-head>` + body + `</mjml>`)
			if err != nil {
				t.Fatal(err)
			}
			if got := mediaQuery.FindAllString(html, -1); !slices.Equal(got, tc.want) {
				t.Errorf("media queries = %q, want %q", got, tc.want)
			}
		})
	}
}
