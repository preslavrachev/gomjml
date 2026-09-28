package mjml

import (
	"strings"
	"testing"
)

// MJML 4.16.1 emits each content as written. Its minified goldens would also
// quote src=x and rewrite nowrap as nowrap="nowrap", so these cases are kept
// out of the fixtures.
func TestEndingTagContentIsRenderedAsWritten(t *testing.T) {
	section := func(content string) string {
		return `<mjml><mj-body><mj-section><mj-column>` + content + `</mj-column></mj-section></mj-body></mjml>`
	}
	tests := []struct {
		name string
		mjml string
		want string
	}{
		{
			name: "mj-button",
			mjml: section(`<mj-button href="https://example.com"><b>Renew</b><br>today <img src=x alt=pic> <span class=hi><i>now</i></span></mj-button>`),
			want: `"><b>Renew</b><br>today <img src=x alt=pic> <span class=hi><i>now</i></span></a>`,
		},
		{
			name: "mj-button keeps br free of an XHTML slash",
			mjml: section(`<mj-button>Line1<br/>Line2<br />Line3</mj-button>`),
			want: `>Line1<br>Line2<br>Line3</p>`,
		},
		{
			name: "mj-table",
			mjml: section(`<mj-table><tr><td nowrap>a<br>b</td><td align=center><img src=x></td></tr></mj-table>`),
			want: `><tr><td nowrap>a<br>b</td><td align=center><img src=x></td></tr></table>`,
		},
		{
			name: "mj-table keeps text and elements in order",
			mjml: section(`<mj-table><tr><td>a <b>b</b> c</td></tr></mj-table>`),
			want: `<td>a <b>b</b> c</td>`,
		},
		{
			name: "mj-accordion-title and mj-accordion-text",
			mjml: section(`<mj-accordion><mj-accordion-element><mj-accordion-title><b>Q</b><br>two</mj-accordion-title><mj-accordion-text><p class="a">a<br>b</p></mj-accordion-text></mj-accordion-element></mj-accordion>`),
			want: `padding:16px;"><b>Q</b><br>two</td>`,
		},
		{
			name: "mj-accordion-text keeps attributes",
			mjml: section(`<mj-accordion><mj-accordion-element><mj-accordion-title>Q</mj-accordion-title><mj-accordion-text><p class="a">a<br>b</p></mj-accordion-text></mj-accordion-element></mj-accordion>`),
			want: `padding:16px;"><p class="a">a<br>b</p></td>`,
		},
		{
			name: "mj-social-element",
			mjml: section(`<mj-social><mj-social-element name="facebook" href="https://example.com"><b>Share</b><br>now</mj-social-element></mj-social>`),
			want: `"><b>Share</b><br>now</a>`,
		},
		{
			name: "mj-navbar-link",
			mjml: section(`<mj-navbar><mj-navbar-link href="/a"><b>Home</b><br>page</mj-navbar-link></mj-navbar>`),
			want: `><b>Home</b><br>page</a>`,
		},
		{
			name: "mj-title",
			mjml: `<mjml><mj-head><mj-title>Title <b>x</b></mj-title></mj-head><mj-body></mj-body></mjml>`,
			want: `<title>Title <b>x</b></title>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			html, err := Render(tt.mjml)
			if err != nil {
				t.Fatalf("Render: %v", err)
			}
			if !strings.Contains(html, tt.want) {
				t.Errorf("output does not contain %s\n%s", tt.want, html)
			}
		})
	}
}
