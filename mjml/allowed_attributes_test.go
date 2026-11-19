package mjml

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

// metadataTemplate is the document from issue #41: metadata kept in an
// attribute MJML does not know.
const metadataTemplate = `<mjml>
 <mj-body>
   <mj-section>
     <mj-column>
       <mj-text custom-attr="some-value" font-family="Helvetica" color="#F45E43">
         <h1>Title</h1>
       </mj-text>
     </mj-column>
   </mj-section>
 </mj-body>
</mjml>`

func TestUnknownAttributeIsReportedByDefault(t *testing.T) {
	html, err := Render(metadataTemplate)

	var mjmlErr Error
	if !errors.As(err, &mjmlErr) {
		t.Fatalf("want an mjml.Error, got %v", err)
	}
	if len(mjmlErr.Details) != 1 || mjmlErr.Details[0].Line != 5 ||
		!strings.Contains(mjmlErr.Details[0].Message, "custom-attr") {
		t.Fatalf("unexpected details: %+v", mjmlErr.Details)
	}
	if !strings.Contains(html, "Title") {
		t.Fatal("the HTML should still be returned alongside the validation error")
	}
}

func TestWithAllowedAttributesAcceptsExtraAttributes(t *testing.T) {
	want, _ := Render(metadataTemplate)

	got, err := Render(metadataTemplate, WithAllowedAttributes(func(tagName, attrName string) bool {
		return strings.HasPrefix(attrName, "custom-")
	}))
	if err != nil {
		t.Fatalf("an accepted attribute must not fail the render: %v", err)
	}
	if got != want {
		t.Fatal("accepting an attribute must not change the HTML")
	}
}

func TestWithAllowedAttributesIsAskedOnlyAboutRejectedAttributes(t *testing.T) {
	src := `<mjml><mj-body><mj-section><mj-column>
<mj-text custom-attr="x" color="red" bogus="1">Hi</mj-text>
</mj-column></mj-section></mj-body></mjml>`

	var asked, reported []string
	_, err := Render(src,
		WithAllowedAttributes(func(tagName, attrName string) bool {
			asked = append(asked, tagName+" "+attrName)
			return attrName == "custom-attr"
		}),
		func(opts *RenderOpts) {
			opts.InvalidAttributeReporter = func(tagName, attrName string, line int) {
				reported = append(reported, tagName+" "+attrName)
			}
		},
	)

	if !slices.Equal(asked, []string{"mj-text custom-attr", "mj-text bogus"}) {
		t.Fatalf("asked about %v", asked)
	}
	if !slices.Equal(reported, []string{"mj-text bogus"}) {
		t.Fatalf("reported %v", reported)
	}
	var mjmlErr Error
	if !errors.As(err, &mjmlErr) || len(mjmlErr.Details) != 1 ||
		!strings.Contains(mjmlErr.Details[0].Message, "bogus") {
		t.Fatalf("want only bogus in the error, got %v", err)
	}
}
