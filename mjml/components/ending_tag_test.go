package components

import "testing"

// The expected values are html-minifier 4.0.0's output for "<div>" + in +
// "</div>" with the options MJML's CLI minifies with (collapseWhitespace,
// caseSensitive, removeEmptyAttributes, minifyCSS off), minus the <div>.
func TestCollapseHTMLWhitespaceMatchesHTMLMinifier(t *testing.T) {
	tests := []struct{ in, want string }{
		{"a <b> b </b> c", "a <b>b </b>c"},
		{"<span>\n   Because emails\n   need tabs.\n   </span>", "<span>Because emails need tabs.</span>"},
		{"<tr>\n  <td>a</td>\n  <td>b</td>\n</tr>", "<tr><td>a</td><td>b</td></tr>"},
		{"a\n<br>\nb", "a<br>b"},
		{"x <!-- c --> y", "x<!-- c --> y"},
		{"<b>Home </b>", "<b>Home</b>"},
		{" \u00a0 x \u00a0 ", "\u00a0 x \u00a0"},
		{"a \n\u00a0\n b", "a \u00a0 b"},
		{"<td> \u00a0 a </td>", "<td>\u00a0 a</td>"},
		{"<pre>  a \n b </pre> c", "<pre>  a \n b </pre>c"},
		{"<textarea>  a \n b </textarea> c", "<textarea>  a \n b </textarea> c"},
		{"x <script>\n var a = 1;  \n</script> y", "x<script>var a = 1;</script>y"},
		{"a <img src=\"x\"> b", "a <img src=\"x\"> b"},
		{"a<wbr> b", "a<wbr> b"},
		{"<nobr>a</nobr> b", "<nobr>a</nobr> b"},
		{"<p>a</p>\n\n<p>b</p>", "<p>a</p><p>b</p>"},
		{"<B> x </B> y", "<B>x</B>y"},
		{"<i>a </i> <i> b</i>", "<i>a </i><i>b</i>"},
	}
	for _, tt := range tests {
		if got := collapseHTMLWhitespace(tt.in); got != tt.want {
			t.Errorf("collapseHTMLWhitespace(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCollapseHTMLWhitespaceKeepsTagsAsWritten(t *testing.T) {
	in := "<td nowrap>a</td>\n<td align=center><img src=x></td><td title='a > b'>c</td>"
	want := "<td nowrap>a</td><td align=center><img src=x></td><td title='a > b'>c</td>"
	if got := collapseHTMLWhitespace(in); got != want {
		t.Errorf("collapseHTMLWhitespace(%q) = %q, want %q", in, got, want)
	}
}
