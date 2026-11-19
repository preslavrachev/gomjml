package parser

import "testing"

func findNode(n *MJMLNode, tag string) *MJMLNode {
	if n.GetTagName() == tag {
		return n
	}
	for _, child := range n.Children {
		if found := findNode(child, tag); found != nil {
			return found
		}
	}
	return nil
}

func TestParseMJMLKeepsEndingTagContentAsWritten(t *testing.T) {
	tests := []struct {
		tag     string
		content string
	}{
		{"mj-text", `<b>Renew</b><br>today <img src=x alt=pic>`},
		{"mj-button", `<b>Renew</b><br>today <img src=x alt=pic> <span class=hi><i>now</i></span>`},
		{"mj-table", `<tr><td nowrap>a<br>b</td><td align=center><img src=x></td></tr>`},
		{"mj-navbar-link", `<b>Home</b><br>page`},
		{"mj-accordion-title", `<b>Q</b><br>two`},
		{"mj-accordion-text", `<p>a<br>b</p> <span class=x>c</span>`},
		{"mj-social-element", `<b>Share</b><br>now`},
		{"mj-raw", `<p>a<br>b</p>`},
		{"mj-title", `Title <b>x</b>`},
		{"mj-preview", `Preview <b>y</b>`},
		{"mj-style", `.a > p { color: red; }`},
	}
	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			root, err := ParseMJML(`<mjml><mj-body><` + tt.tag + ` css-class="c">` + tt.content + `</` + tt.tag + `></mj-body></mjml>`)
			if err != nil {
				t.Fatalf("ParseMJML: %v", err)
			}
			node := findNode(root, tt.tag)
			if node == nil {
				t.Fatalf("no <%s> in the AST", tt.tag)
			}
			if node.Text != tt.content {
				t.Errorf("Text = %q, want %q", node.Text, tt.content)
			}
			if len(node.Children) != 0 {
				t.Errorf("content was parsed into %d child nodes", len(node.Children))
			}
			if got := node.GetAttribute("css-class"); got != "c" {
				t.Errorf("css-class = %q, want %q", got, "c")
			}
		})
	}
}

// MJML counts every ending tag while it looks for the one that closes the
// content, and it reads comments and quoted attribute values as they are.
func TestParseMJMLFindsTheEndingTagThatClosesTheContent(t *testing.T) {
	tests := []struct {
		name    string
		mjml    string
		tag     string
		content string
	}{
		{
			name:    "ending tag inside mj-raw",
			mjml:    `<mjml><mj-body><mj-raw><mj-text>x</mj-text></mj-raw></mj-body></mjml>`,
			tag:     "mj-raw",
			content: `<mj-text>x</mj-text>`,
		},
		{
			name:    "closing tag inside a quoted attribute value",
			mjml:    `<mjml><mj-body><mj-button><a title="</mj-button>">x</a></mj-button></mj-body></mjml>`,
			tag:     "mj-button",
			content: `<a title="</mj-button>">x</a>`,
		},
		{
			name:    "closing tag inside a comment",
			mjml:    `<mjml><mj-body><mj-table><!-- </mj-table> --><tr><td>x</td></tr></mj-table></mj-body></mjml>`,
			tag:     "mj-table",
			content: `<!-- </mj-table> --><tr><td>x</td></tr>`,
		},
		{
			name:    "self-closing ending tag inside the content",
			mjml:    `<mjml><mj-body><mj-raw><mj-text /><br></mj-raw></mj-body></mjml>`,
			tag:     "mj-raw",
			content: `<mj-text /><br>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root, err := ParseMJML(tt.mjml)
			if err != nil {
				t.Fatalf("ParseMJML: %v", err)
			}
			if node := findNode(root, tt.tag); node == nil || node.Text != tt.content {
				t.Fatalf("<%s> content = %+v, want %q", tt.tag, node, tt.content)
			}
		})
	}
}
