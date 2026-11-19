package mjml

import (
	"encoding/json"
	"fmt"
	"maps"
	"math/rand/v2"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// diffCase is one input of the differential corpus.
type diffCase struct {
	Name     string // "<source>/<name>", unique across the corpus
	MJML     string
	Pkg      string // package.json alias of mjml that renders the reference; "" is "mjml"
	FilePath string // resolves mj-include on the reference side
}

// fuzzSeeds is how many grammar fuzzer documents the corpus compares with MJML.
const fuzzSeeds = 400

const differentialDir = "testdata/differential"

// differentialCache holds what scripts/differential.sh fetches and renders: the
// email-templates checkout and the reference results. It is not committed.
var differentialCache = filepath.Join(differentialDir, ".cache")

// testdataCases are the fixtures of TestMJMLAgainstExpected, with their committed goldens.
func testdataCases() ([]diffCase, error) {
	inputs, err := filepath.Glob("testdata/*.mjml")
	if err != nil {
		return nil, err
	}
	// Matches the overrides in testdata/reference/render.mjs.
	overrides := map[string]string{"mj-wrapper-gap": "mjml-4.17"}
	var cases []diffCase
	for _, input := range inputs {
		name := strings.TrimSuffix(filepath.Base(input), ".mjml")
		content, err := os.ReadFile(input)
		if err != nil {
			return nil, err
		}
		cases = append(cases, diffCase{Name: "testdata/" + name, MJML: string(content), Pkg: overrides[name]})
	}
	return cases, nil
}

// directoryCases reads every .mjml file of dir as a case of source.
func directoryCases(source, dir string) ([]diffCase, error) {
	inputs, err := filepath.Glob(filepath.Join(dir, "*.mjml"))
	if err != nil {
		return nil, err
	}
	if len(inputs) == 0 {
		return nil, errSourceUnavailable
	}
	var cases []diffCase
	for _, input := range inputs {
		content, err := os.ReadFile(input)
		if err != nil {
			return nil, err
		}
		absolute, _ := filepath.Abs(input)
		cases = append(cases, diffCase{
			Name:     source + "/" + strings.TrimSuffix(filepath.Base(input), ".mjml"),
			MJML:     string(content),
			FilePath: absolute,
		})
	}
	return cases, nil
}

// componentAttributes reads the attribute types gomjml allows per component. Callers share
// the result and must not modify it.
var componentAttributes = sync.OnceValues(func() (map[string]map[string]string, error) {
	content, err := os.ReadFile("components/allowed-css-attributes.json")
	if err != nil {
		return nil, err
	}
	var attributes map[string]map[string]string
	if err := json.Unmarshal(content, &attributes); err != nil {
		return nil, err
	}
	// allowed-css-attributes.json has no mj-wrapper; MJML's mj-wrapper extends mj-section.
	attributes["mj-wrapper"] = maps.Clone(attributes["mj-section"])
	for component, set := range attributes {
		if !headComponents[component] {
			set["css-class"] = "string"
		}
	}
	return attributes, nil
})

var headComponents = map[string]bool{"mj-breakpoint": true, "mj-font": true, "mj-style": true, "mj-raw": true}

// stringValues are representative values for attributes typed "string", by attribute name.
var stringValues = map[string][]string{
	"alt":                   {"An 'alt' &amp; more"},
	"background-position":   {"top left", "center center", "10% 20%"},
	"background-position-x": {"left", "30%"},
	"background-position-y": {"bottom", "30%"},
	"background-size":       {"cover", "100px auto"},
	"background-url":        {"https://example.com/bg.png"},
	"base-url":              {"https://example.com/"},
	"border":                {"2px dashed #ff0000", "none"},
	"border-radius":         {"8px", "4px 8px", "50%"},
	"border-style":          {"dashed", "dotted"},
	"css-class":             {"custom-class"},
	"font-family":           {"Georgia, serif", "'Open Sans', Arial"},
	"font-style":            {"italic"},
	"font-weight":           {"bold", "300"},
	"hamburger":             {"hamburger"},
	"href":                  {"https://example.com/path?a=1&amp;b=2"},
	"ico-close":             {"&#10005;"},
	"ico-font-family":       {"Arial"},
	"ico-open":              {"&#9776;"},
	"ico-text-decoration":   {"underline"},
	"ico-text-transform":    {"lowercase"},
	"inline":                {"inline"},
	"left-icon":             {"https://example.com/left.png"},
	"mode":                  {"fixed-height", "fluid-height"},
	"name":                  {"facebook", "custom"},
	"rel":                   {"noopener"},
	"right-icon":            {"https://example.com/right.png"},
	"sizes":                 {"(max-width: 600px) 100vw, 600px"},
	"src":                   {"https://example.com/img.png?w=300&amp;h=200"},
	"srcset":                {"https://example.com/a.png 1x, https://example.com/b.png 2x"},
	"target":                {"_self"},
	"tb-border":             {"2px solid #336699"},
	"text-decoration":       {"underline", "none"},
	"text-transform":        {"uppercase", "capitalize"},
	"thumbnails-src":        {"https://example.com/thumb.png"},
	"title":                 {"A title"},
	"usemap":                {"#map"},
}

var enumPattern = regexp.MustCompile(`^enum\((.*)\)$`)

// representativeValues lists values for an attribute of the given allowed-css-attributes type.
func representativeValues(attribute, kind string) []string {
	if strings.HasPrefix(attribute, "border") && kind == "string" && attribute != "border-radius" && attribute != "border-style" {
		return stringValues["border"]
	}
	if strings.HasPrefix(attribute, "inner-border") {
		return []string{"1px solid #336699"}
	}
	if strings.HasPrefix(attribute, "icon-") && strings.HasSuffix(attribute, "-url") {
		return []string{"https://example.com/icon.png"}
	}
	if strings.HasPrefix(attribute, "icon-") && strings.HasSuffix(attribute, "-alt") {
		return []string{"toggle"}
	}
	switch kind {
	case "color":
		return []string{"#336699", "#f00", "red", "rgba(255, 0, 0, 0.5)"}
	case "unit(px,%)":
		return []string{"0px", "12px", "50%"}
	case "unit(px,%){1,4}":
		return []string{"0", "10px", "10px 20px", "1px 2px 3px", "1px 2px 3px 4px", "5%"}
	case "unit(px)":
		return []string{"0px", "25px"}
	case "unit(px,%,)":
		return []string{"20px", "150%", "1.5"}
	case "unitWithNegative(px,em)":
		return []string{"-1px", "0.2em"}
	case "unit(px,auto)":
		return []string{"100px", "auto"}
	case "unit(px,%,auto)":
		return []string{"100px", "50%", "auto"}
	case "integer":
		return []string{"0", "4"}
	case "boolean":
		return []string{"true", "false"}
	case "string":
		if values, ok := stringValues[attribute]; ok {
			return values
		}
		if strings.Contains(attribute, "width") || strings.Contains(attribute, "height") {
			return []string{"100px", "50%"}
		}
		return []string{"value"}
	}
	if match := enumPattern.FindStringSubmatch(kind); match != nil {
		var values []string
		for value := range strings.SplitSeq(match[1], ",") {
			if value != "" {
				values = append(values, value)
			}
		}
		return values
	}
	return []string{"value"}
}

// mjmlDocument assembles a document from head and body markup.
func mjmlDocument(head, bodyAttributes, body string) string {
	var b strings.Builder
	b.WriteString("<mjml>\n")
	if head != "" {
		b.WriteString("  <mj-head>\n" + head + "\n  </mj-head>\n")
	}
	b.WriteString("  <mj-body" + bodyAttributes + ">\n" + body + "\n  </mj-body>\n</mjml>\n")
	return b.String()
}

func inColumn(content string) string {
	return "    <mj-section>\n      <mj-column>\n        " + content + "\n      </mj-column>\n    </mj-section>"
}

// componentSkeleton returns a minimal valid document that exercises component, with attrs on it.
func componentSkeleton(component, attrs, head string) string {
	text := `<mj-text>Some text</mj-text>`
	switch component {
	case "mj-body":
		return mjmlDocument(head, attrs, inColumn(text))
	case "mj-breakpoint":
		return mjmlDocument(head+"    <mj-breakpoint"+attrs+" />", "",
			"    <mj-section>\n      <mj-column><mj-text>Left</mj-text></mj-column>\n      <mj-column><mj-text>Right</mj-text></mj-column>\n    </mj-section>")
	case "mj-font":
		return mjmlDocument(head+"    <mj-font"+withDefaults(` name="Raleway" href="https://fonts.googleapis.com/css?family=Raleway"`, attrs)+" />", "",
			inColumn(`<mj-text font-family="Raleway, Arial">Some text</mj-text>`))
	case "mj-style":
		return mjmlDocument(head+"    <mj-style"+attrs+">.red { color: red; } .pad td { padding: 2px; }</mj-style>", "",
			inColumn(`<mj-text css-class="red">Some text</mj-text>`))
	case "mj-raw":
		return mjmlDocument(head, "", inColumn(text)+"\n    <mj-raw"+attrs+"><p>raw</p></mj-raw>")
	case "mj-section":
		return mjmlDocument(head, "", "    <mj-section"+attrs+">\n      <mj-column>"+text+"</mj-column>\n    </mj-section>")
	case "mj-wrapper":
		return mjmlDocument(head, "", "    <mj-wrapper"+attrs+">\n      <mj-section><mj-column>"+text+"</mj-column></mj-section>\n    </mj-wrapper>")
	case "mj-column":
		return mjmlDocument(head, "", "    <mj-section>\n      <mj-column"+attrs+">"+text+"</mj-column>\n      <mj-column>"+text+"</mj-column>\n    </mj-section>")
	case "mj-group":
		return mjmlDocument(head, "", "    <mj-section>\n      <mj-group"+attrs+">\n        <mj-column>"+text+"</mj-column>\n        <mj-column>"+text+"</mj-column>\n      </mj-group>\n    </mj-section>")
	case "mj-hero":
		return mjmlDocument(head, "", "    <mj-hero"+withDefaults(` background-url="https://example.com/hero.jpg" background-height="400px" background-width="600px"`, attrs)+">\n      "+text+"\n    </mj-hero>")
	case "mj-button":
		return mjmlDocument(head, "", inColumn("<mj-button"+withDefaults(` href="https://example.com"`, attrs)+">Click</mj-button>"))
	case "mj-text":
		return mjmlDocument(head, "", inColumn("<mj-text"+attrs+">Some text</mj-text>"))
	case "mj-image":
		return mjmlDocument(head, "", inColumn("<mj-image"+withDefaults(` src="https://example.com/img.png"`, attrs)+" />"))
	case "mj-divider", "mj-spacer":
		return mjmlDocument(head, "", inColumn("<"+component+attrs+" />"))
	case "mj-table":
		return mjmlDocument(head, "", inColumn("<mj-table"+attrs+"><tr><th>A</th><th>B</th></tr><tr><td>1</td><td>2</td></tr></mj-table>"))
	case "mj-social":
		return mjmlDocument(head, "", inColumn("<mj-social"+attrs+`><mj-social-element name="facebook" href="https://facebook.com">Facebook</mj-social-element><mj-social-element name="twitter" href="https://twitter.com" /></mj-social>`))
	case "mj-social-element":
		return mjmlDocument(head, "", inColumn("<mj-social><mj-social-element"+withDefaults(` name="facebook" href="https://facebook.com"`, attrs)+`>Facebook</mj-social-element><mj-social-element name="twitter" /></mj-social>`))
	case "mj-navbar":
		return mjmlDocument(head, "", inColumn("<mj-navbar"+attrs+`><mj-navbar-link href="/a">A</mj-navbar-link><mj-navbar-link href="/b">B</mj-navbar-link></mj-navbar>`))
	case "mj-navbar-link":
		return mjmlDocument(head, "", inColumn(`<mj-navbar base-url="https://example.com"><mj-navbar-link`+withDefaults(` href="/a"`, attrs)+`>A</mj-navbar-link><mj-navbar-link href="/b">B</mj-navbar-link></mj-navbar>`))
	case "mj-carousel":
		return mjmlDocument(head, "", inColumn("<mj-carousel"+attrs+`><mj-carousel-image src="https://example.com/1.png" /><mj-carousel-image src="https://example.com/2.png" /></mj-carousel>`))
	case "mj-carousel-image":
		return mjmlDocument(head, "", inColumn("<mj-carousel><mj-carousel-image"+withDefaults(` src="https://example.com/1.png"`, attrs)+` /><mj-carousel-image src="https://example.com/2.png" /></mj-carousel>`))
	case "mj-accordion", "mj-accordion-element", "mj-accordion-title", "mj-accordion-text":
		on := func(tag string) string {
			if tag == component {
				return attrs
			}
			return ""
		}
		return mjmlDocument(head, "", inColumn("<mj-accordion"+on("mj-accordion")+">"+
			"<mj-accordion-element"+on("mj-accordion-element")+">"+
			"<mj-accordion-title"+on("mj-accordion-title")+">Title</mj-accordion-title>"+
			"<mj-accordion-text"+on("mj-accordion-text")+">Body</mj-accordion-text>"+
			"</mj-accordion-element><mj-accordion-element><mj-accordion-title>Second</mj-accordion-title><mj-accordion-text>More</mj-accordion-text></mj-accordion-element></mj-accordion>"))
	}
	panic("no skeleton for " + component)
}

var attributeName = regexp.MustCompile(` ([a-z-]+)="`)

// withDefaults returns a skeleton's default attributes, leaving out those attrs sets, then attrs:
// a tag with one attribute twice is not XML, and MJML and gomjml keep different ones.
func withDefaults(defaults, attrs string) string {
	for _, match := range attributeName.FindAllStringSubmatch(attrs, -1) {
		defaults = regexp.MustCompile(` `+match[1]+`="[^"]*"`).ReplaceAllString(defaults, "")
	}
	return defaults + attrs
}

var slugUnsafe = regexp.MustCompile(`[^a-zA-Z0-9.%-]+`)

func slug(value string) string {
	if value == "" {
		return "empty"
	}
	return strings.Trim(slugUnsafe.ReplaceAllString(value, "_"), "_")
}

// generatedCases builds every component × attribute × representative value in a minimal
// skeleton, each attribute's first value again through mj-attributes, the nesting
// combinations of the layout components, and content cases.
func generatedCases() ([]diffCase, error) {
	attributes, err := componentAttributes()
	if err != nil {
		return nil, err
	}
	var cases []diffCase
	seen := map[string]bool{}
	add := func(name, mjml string) {
		name = "generated/" + name
		for base, i := name, 2; seen[name]; i++ {
			name = fmt.Sprintf("%s~%d", base, i)
		}
		seen[name] = true
		cases = append(cases, diffCase{Name: name, MJML: mjml})
	}

	components := slices.Sorted(maps.Keys(attributes))
	for _, component := range components {
		// The skeleton alone, so that clustering can tell its differences from an attribute's.
		add("baseline/"+component, componentSkeleton(component, "", ""))
		names := slices.Sorted(maps.Keys(attributes[component]))
		for _, attribute := range names {
			values := representativeValues(attribute, attributes[component][attribute])
			for _, value := range values {
				attrs := fmt.Sprintf(` %s="%s"`, attribute, value)
				add(fmt.Sprintf("attr/%s/%s=%s", component, attribute, slug(value)), componentSkeleton(component, attrs, ""))
			}
			if !headComponents[component] && component != "mj-body" {
				head := fmt.Sprintf(`    <mj-attributes><%s %s="%s" /></mj-attributes>`+"\n", component, attribute, values[0])
				add(fmt.Sprintf("mj-attributes/%s/%s=%s", component, attribute, slug(values[0])), componentSkeleton(component, "", head))
			}
		}
		if !headComponents[component] && component != "mj-body" {
			head := `    <mj-attributes><mj-class name="c1" padding="7px" color="#123456" /><mj-all font-family="Verdana" /></mj-attributes>` + "\n"
			add("mj-class/"+component, componentSkeleton(component, ` mj-class="c1"`, head))
		}
	}

	for _, n := range nestingCases() {
		add("nesting/"+n.name, n.mjml)
	}
	for _, c := range contentCases {
		add("content/"+c.name, c.mjml)
	}
	return cases, nil
}

type namedDocument struct{ name, mjml string }

// nestingCases combines the layout containers with column counts and widths.
func nestingCases() []namedDocument {
	containers := []struct {
		name        string
		open, close string
	}{
		{"section", "<mj-section>", "</mj-section>"},
		{"section-full-width", `<mj-section full-width="full-width" background-color="#eee">`, "</mj-section>"},
		{"section-rtl", `<mj-section direction="rtl">`, "</mj-section>"},
		{"wrapper-section", `<mj-wrapper padding="10px" background-color="#ddd"><mj-section>`, "</mj-section></mj-wrapper>"},
		{"wrapper-full-width-section", `<mj-wrapper full-width="full-width" background-url="https://example.com/bg.png"><mj-section>`, "</mj-section></mj-wrapper>"},
		{"section-group", "<mj-section><mj-group>", "</mj-group></mj-section>"},
		{"wrapper-section-group", `<mj-wrapper border="1px solid #000"><mj-section><mj-group direction="rtl">`, "</mj-group></mj-section></mj-wrapper>"},
	}
	widths := []struct {
		name  string
		width func(i, n int) string
	}{
		{"auto", func(int, int) string { return "" }},
		{"percent", func(i, n int) string {
			if i == 0 && n > 1 {
				return ` width="40%"`
			}
			return fmt.Sprintf(` width="%d%%"`, 60/max(n-1, 1))
		}},
		{"pixels", func(i, n int) string { return fmt.Sprintf(` width="%dpx"`, 600/n-10*i) }},
	}
	contents := []string{
		`<mj-text>Text</mj-text>`,
		`<mj-image src="https://example.com/img.png" />`,
		`<mj-button href="#">Go</mj-button>`,
		`<mj-divider />`,
	}
	var docs []namedDocument
	for _, container := range containers {
		for n := 1; n <= 4; n++ {
			for _, width := range widths {
				if width.name != "auto" && n == 1 && container.name != "section" {
					continue
				}
				var columns strings.Builder
				for i := range n {
					fmt.Fprintf(&columns, "<mj-column%s>%s</mj-column>", width.width(i, n), contents[(i+n)%len(contents)])
				}
				body := "    " + container.open + columns.String() + container.close
				docs = append(docs, namedDocument{fmt.Sprintf("%s/%d-columns-%s", container.name, n, width.name), mjmlDocument("", "", body)})
			}
		}
	}
	heroes := []string{
		`<mj-hero mode="fixed-height" height="300px" background-url="https://example.com/h.jpg" background-width="600px" background-height="300px"><mj-text>Hero</mj-text><mj-button>Go</mj-button></mj-hero>`,
		`<mj-hero mode="fluid-height" background-url="https://example.com/h.jpg" background-width="600px" background-height="300px" padding="20px"><mj-image src="https://example.com/i.png" /><mj-text>Hero</mj-text></mj-hero>`,
		`<mj-wrapper><mj-hero background-color="#333"><mj-text color="#fff">In wrapper</mj-text></mj-hero></mj-wrapper>`,
		`<mj-hero background-url="https://example.com/h.jpg" background-position="top left"><mj-spacer height="20px" /><mj-divider /><mj-table><tr><td>t</td></tr></mj-table></mj-hero>`,
	}
	for i, hero := range heroes {
		docs = append(docs, namedDocument{fmt.Sprintf("hero/%d", i+1), mjmlDocument("", "", "    "+hero)})
	}
	mixed := []string{
		`<mj-section><mj-column width="30%"><mj-text>A</mj-text></mj-column><mj-group width="70%"><mj-column><mj-text>B</mj-text></mj-column><mj-column><mj-text>C</mj-text></mj-column></mj-group></mj-section>`,
		`<mj-wrapper><mj-section><mj-column><mj-text>1</mj-text></mj-column></mj-section><mj-section><mj-column><mj-text>2</mj-text></mj-column><mj-column><mj-text>3</mj-text></mj-column></mj-section></mj-wrapper>`,
		`<mj-section><mj-column></mj-column></mj-section>`,
		`<mj-section></mj-section>`,
		`<mj-wrapper></mj-wrapper>`,
		`<mj-section padding="0"><mj-column padding="10px" border="1px solid red" background-color="#eee" inner-background-color="#fff"><mj-text>Padded</mj-text></mj-column></mj-section>`,
		`<mj-raw><div>before</div></mj-raw><mj-section><mj-column><mj-raw><span>inside</span></mj-raw><mj-text>x</mj-text></mj-column></mj-section><mj-raw><div>after</div></mj-raw>`,
	}
	for i, body := range mixed {
		docs = append(docs, namedDocument{fmt.Sprintf("mixed/%d", i+1), mjmlDocument("", "", "    "+body)})
	}
	return docs
}

// contentCases exercise text handling: whitespace, entities, comments, void tags and conditionals.
var contentCases = []namedDocument{
	{"pre-blank-lines", mjmlDocument("", "", inColumn("<mj-text><pre>line one\n\nline three\n\n\n  indented</pre></mj-text>"))},
	{"pre-in-raw", mjmlDocument("", "", "    <mj-raw><pre>raw\n\n  pre</pre></mj-raw>")},
	{"white-space-pre", mjmlDocument("", "", inColumn("<mj-text><div style=\"white-space:pre\">a\n\n  b   c</div></mj-text>"))},
	{"white-space-pre-wrap", mjmlDocument("", "", inColumn("<mj-text><p style=\"white-space: pre-wrap\">a  b\n  c</p></mj-text>"))},
	{"white-space-nowrap", mjmlDocument("", "", inColumn("<mj-text><span style=\"white-space:nowrap\">no   wrap here</span></mj-text>"))},
	{"text-multiline", mjmlDocument("", "", inColumn("<mj-text>\n  first line\n\n  second   line\n</mj-text>"))},
	{"entities-named", mjmlDocument("", "", inColumn("<mj-text>&nbsp;&amp;&lt;&gt;&quot;&copy;&eacute;&mdash;</mj-text>"))},
	{"entities-numeric", mjmlDocument("", "", inColumn("<mj-text>&#169; &#xA9; &#8212; &#x1F600;</mj-text>"))},
	{"entities-in-attribute", mjmlDocument("", "", inColumn(`<mj-button href="https://example.com/?a=1&amp;b=2&amp;c=&quot;x&quot;" title="&lt;t&gt;">Go &amp; see</mj-button>`))},
	{"entities-in-button", mjmlDocument("", "", inColumn("<mj-button>Save&nbsp;now &rarr;</mj-button>"))},
	{"unicode", mjmlDocument("", "", inColumn("<mj-text>Grüße — 日本語 — emoji 🎉</mj-text>"))},
	{"comment-in-text", mjmlDocument("", "", inColumn("<mj-text>before <!-- a comment --> after</mj-text>"))},
	{"comment-between-components", mjmlDocument("", "", "    <!-- top -->\n    <mj-section>\n      <!-- in section -->\n      <mj-column>\n        <!-- in column -->\n        <mj-text>x</mj-text>\n      </mj-column>\n    </mj-section>")},
	{"comment-in-head", "<mjml>\n  <mj-head>\n    <!-- head comment -->\n    <mj-title>T</mj-title>\n  </mj-head>\n  <mj-body>" + inColumn("<mj-text>x</mj-text>") + "</mj-body>\n</mjml>\n"},
	{"void-br", mjmlDocument("", "", inColumn("<mj-text>a<br>b<br/>c<br />d</mj-text>"))},
	{"void-img-hr", mjmlDocument("", "", inColumn(`<mj-text><img src="https://example.com/i.png" alt="i"><hr><input type="text"></mj-text>`))},
	{"void-in-raw", mjmlDocument("", "", `    <mj-raw><img src="https://example.com/i.png"><br><meta name="x" content="y"></mj-raw>`)},
	{"mso-in-text", mjmlDocument("", "", inColumn("<mj-text><!--[if mso]><b>Outlook</b><![endif]--><!--[if !mso]><!--><i>Others</i><!--<![endif]--></mj-text>"))},
	{"mso-in-raw", mjmlDocument("", "", "    <mj-raw><!--[if mso | IE]><table><tr><td><![endif]--><div>content</div><!--[if mso | IE]></td></tr></table><![endif]--></mj-raw>")},
	{"mso-in-head-raw", "<mjml>\n  <mj-head>\n    <mj-raw><!--[if mso]><style>.x{color:red}</style><![endif]--></mj-raw>\n  </mj-head>\n  <mj-body>" + inColumn("<mj-text>x</mj-text>") + "</mj-body>\n</mjml>\n"},
	{"inline-markup", mjmlDocument("", "", inColumn(`<mj-text><p>Para <b>bold</b> <em>em</em> <a href="https://example.com" style="color:red">link</a></p><ul><li>one</li><li>two</li></ul></mj-text>`))},
	{"nested-tables-in-text", mjmlDocument("", "", inColumn("<mj-text><table><tr><td>cell</td></tr></table></mj-text>"))},
	{"table-content", mjmlDocument("", "", inColumn(`<mj-table><tr style="border-bottom:1px solid #ecedee;text-align:left;"><th>Year</th><th>Language</th></tr><tr><td>1995</td><td>PHP &amp; JS</td></tr></mj-table>`))},
	{"button-html", mjmlDocument("", "", inColumn("<mj-button><b>Bold</b> <span style=\"color:red\">red</span></mj-button>"))},
	{"social-text-html", mjmlDocument("", "", inColumn(`<mj-social><mj-social-element name="facebook"><b>Fb</b> page</mj-social-element></mj-social>`))},
	{"navbar-text-html", mjmlDocument("", "", inColumn(`<mj-navbar><mj-navbar-link href="/a"><b>A</b> &amp; B</mj-navbar-link></mj-navbar>`))},
	{"accordion-html", mjmlDocument("", "", inColumn("<mj-accordion><mj-accordion-element><mj-accordion-title><b>T</b></mj-accordion-title><mj-accordion-text><p>P1</p><p>P2</p></mj-accordion-text></mj-accordion-element></mj-accordion>"))},
	{"empty-text", mjmlDocument("", "", inColumn("<mj-text></mj-text>"))},
	{"whitespace-only-text", mjmlDocument("", "", inColumn("<mj-text>   \n   </mj-text>"))},
	{"head-full", "<mjml>\n  <mj-head>\n    <mj-title>Title &amp; more</mj-title>\n    <mj-preview>Preview text</mj-preview>\n    <mj-breakpoint width=\"320px\" />\n    <mj-font name=\"Roboto\" href=\"https://fonts.googleapis.com/css?family=Roboto\" />\n    <mj-style>.a { color: red; }</mj-style>\n    <mj-style inline=\"inline\">.b { color: blue; }</mj-style>\n    <mj-attributes><mj-all font-family=\"Roboto, Arial\" /><mj-text color=\"#555\" /></mj-attributes>\n  </mj-head>\n  <mj-body>" + inColumn(`<mj-text css-class="a">A</mj-text><mj-text css-class="b">B</mj-text>`) + "</mj-body>\n</mjml>\n"},
	{"closing-tag-whitespace", "<mjml>\n  <mj-body>\n    <mj-section>\n      <mj-column>\n        <mj-text\n          padding=\"0px\"\n          >One</mj-text\n        >\n        <mj-button href=\"#\">Two</mj-button >\n        <mj-text>Three</mj-text>\n      </mj-column\n      >\n    </mj-section>\n  </mj-body>\n</mjml>\n"},
	{"html-in-title-preview", "<mjml>\n  <mj-head>\n    <mj-title>A <i>title</i></mj-title>\n    <mj-preview>A <b>preview</b></mj-preview>\n  </mj-head>\n  <mj-body>" + inColumn("<mj-text>x</mj-text>") + "</mj-body>\n</mjml>\n"},
	{"lang-dir", "<mjml lang=\"fr\" dir=\"rtl\">\n  <mj-body>" + inColumn("<mj-text>Bonjour</mj-text>") + "</mj-body>\n</mjml>\n"},
	{"inline-style-selectors", "<mjml>\n  <mj-head>\n    <mj-style inline=\"inline\">.t div { font-size: 20px !important; } a { color: green; }</mj-style>\n  </mj-head>\n  <mj-body>" + inColumn(`<mj-text css-class="t"><a href="#">link</a></mj-text>`) + "</mj-body>\n</mjml>\n"},
	{"cdata-like-text", mjmlDocument("", "", inColumn("<mj-text>if (a &lt; b &amp;&amp; c &gt; d) {}</mj-text>"))},
	{"long-text", mjmlDocument("", "", inColumn("<mj-text>"+strings.Repeat("Lorem ipsum dolor sit amet. ", 40)+"</mj-text>"))},
}

// fuzzCases are the grammar fuzzer's documents for seeds 1..fuzzSeeds.
func fuzzCases() []diffCase {
	cases := make([]diffCase, 0, fuzzSeeds)
	for seed := uint64(1); seed <= fuzzSeeds; seed++ {
		cases = append(cases, diffCase{Name: fmt.Sprintf("fuzz/seed-%04d", seed), MJML: fuzzDocument(seed)})
	}
	return cases
}

// fuzzer writes random valid MJML from a small grammar. The same seed always gives the same
// document, as math/rand/v2's PCG is specified.
type fuzzer struct {
	r          *rand.Rand
	attributes map[string]map[string]string
	classes    []string
}

func fuzzDocument(seed uint64) string {
	attributes, err := componentAttributes()
	if err != nil {
		panic(err)
	}
	f := &fuzzer{r: rand.New(rand.NewPCG(seed, 0x6d6a6d6c)), attributes: attributes}
	return f.document()
}

func (f *fuzzer) chance(p float64) bool { return f.r.Float64() < p }

func (f *fuzzer) pick(options []string) string { return options[f.r.IntN(len(options))] }

// attrs returns up to n random attributes that component allows, with representative values,
// leaving out the ones the caller already writes.
func (f *fuzzer) attrs(component string, n int, written ...string) string {
	allowed := f.attributes[component]
	names := slices.Sorted(maps.Keys(allowed))
	var b strings.Builder
	used := map[string]bool{}
	for range f.r.IntN(n + 1) {
		name := f.pick(names)
		if used[name] || slices.Contains(written, name) {
			continue
		}
		used[name] = true
		fmt.Fprintf(&b, ` %s="%s"`, name, f.pick(representativeValues(name, allowed[name])))
	}
	if len(f.classes) > 0 && f.chance(0.15) {
		fmt.Fprintf(&b, ` mj-class="%s"`, f.pick(f.classes))
	}
	return b.String()
}

var fuzzWords = []string{"Hello", "world", "&amp;", "&nbsp;", "price: 10&#8364;", "<b>bold</b>", "<i>it</i>", "<br />",
	`<a href="https://example.com/?q=1&amp;r=2">link</a>`, `<span style="color:#f00">red</span>`, "<p>para</p>", "Grüße", "a  b"}

// label is text for components that render it inside a link or a label, where a nested <a> or
// <p> would be invalid HTML that parsers restructure.
func (f *fuzzer) label() string {
	var words []string
	for range 1 + f.r.IntN(3) {
		word := f.pick(fuzzWords)
		if !strings.Contains(word, "<a ") && !strings.Contains(word, "<p>") {
			words = append(words, word)
		}
	}
	if len(words) == 0 {
		return "label"
	}
	return strings.Join(words, " ")
}

func (f *fuzzer) text() string {
	var words []string
	for range 1 + f.r.IntN(6) {
		words = append(words, f.pick(fuzzWords))
	}
	return strings.Join(words, " ")
}

func (f *fuzzer) document() string {
	var head strings.Builder
	if f.chance(0.6) {
		if f.chance(0.5) {
			head.WriteString("<mj-title>" + f.pick([]string{"Hello &amp; welcome", "Title", "Grüße"}) + "</mj-title>")
		}
		if f.chance(0.3) {
			head.WriteString("<mj-preview>Preview " + f.pick([]string{"text", "&amp; more", "price: 10&#8364;"}) + "</mj-preview>")
		}
		if f.chance(0.2) {
			head.WriteString(`<mj-breakpoint width="` + f.pick([]string{"320px", "400px", "600px"}) + `" />`)
		}
		if f.chance(0.2) {
			head.WriteString(`<mj-font name="Lato" href="https://fonts.googleapis.com/css?family=Lato" />`)
		}
		if f.chance(0.4) {
			head.WriteString("<mj-attributes>")
			if f.chance(0.5) {
				head.WriteString(`<mj-all font-family="` + f.pick([]string{"Lato, Arial", "Georgia", "Helvetica"}) + `" />`)
			}
			for range f.r.IntN(3) {
				component := f.pick([]string{"mj-text", "mj-button", "mj-section", "mj-column", "mj-image", "mj-divider"})
				head.WriteString("<" + component + f.attrs(component, 2) + " />")
			}
			for i := range f.r.IntN(3) {
				name := fmt.Sprintf("c%d", i)
				f.classes = append(f.classes, name)
				head.WriteString(`<mj-class name="` + name + `"` + f.attrs("mj-text", 2) + " />")
			}
			head.WriteString("</mj-attributes>")
		}
		if f.chance(0.3) {
			inline := ""
			if f.chance(0.5) {
				inline = ` inline="inline"`
			}
			head.WriteString("<mj-style" + inline + ">.custom-class { color: #123; } a { text-decoration: none; }</mj-style>")
		}
	}
	var body strings.Builder
	for range 1 + f.r.IntN(4) {
		switch x := f.r.IntN(10); {
		case x < 6:
			body.WriteString(f.section())
		case x < 8:
			body.WriteString("<mj-wrapper" + f.attrs("mj-wrapper", 3) + ">")
			for range 1 + f.r.IntN(3) {
				body.WriteString(f.section())
			}
			body.WriteString("</mj-wrapper>")
		case x < 9:
			body.WriteString(`<mj-hero background-url="https://example.com/h.jpg" background-width="600px" background-height="300px"` + f.attrs("mj-hero", 3, "background-url", "background-width", "background-height") + ">")
			for range 1 + f.r.IntN(3) {
				body.WriteString(f.content())
			}
			body.WriteString("</mj-hero>")
		default:
			body.WriteString("<mj-raw><div>" + f.text() + "</div></mj-raw>")
		}
	}
	return mjmlDocument(head.String(), f.attrs("mj-body", 2), body.String())
}

func (f *fuzzer) section() string {
	var b strings.Builder
	b.WriteString("<mj-section" + f.attrs("mj-section", 3) + ">")
	columns := func(n int) {
		for range n {
			b.WriteString("<mj-column" + f.attrs("mj-column", 3) + ">")
			for range 1 + f.r.IntN(3) {
				b.WriteString(f.content())
			}
			b.WriteString("</mj-column>")
		}
	}
	if f.chance(0.2) {
		b.WriteString("<mj-group" + f.attrs("mj-group", 2) + ">")
		columns(1 + f.r.IntN(3))
		b.WriteString("</mj-group>")
	} else {
		columns(1 + f.r.IntN(4))
	}
	b.WriteString("</mj-section>")
	return b.String()
}

func (f *fuzzer) content() string {
	switch f.r.IntN(12) {
	case 0, 1, 2:
		return "<mj-text" + f.attrs("mj-text", 3) + ">" + f.text() + "</mj-text>"
	case 3:
		return `<mj-button href="https://example.com"` + f.attrs("mj-button", 3, "href") + ">" + f.label() + "</mj-button>"
	case 4:
		return `<mj-image src="https://example.com/i.png"` + f.attrs("mj-image", 3, "src") + " />"
	case 5:
		return "<mj-divider" + f.attrs("mj-divider", 3) + " />"
	case 6:
		return "<mj-spacer" + f.attrs("mj-spacer", 2) + " />"
	case 7:
		return "<mj-table" + f.attrs("mj-table", 2) + "><tr><td>" + f.text() + "</td><td>2</td></tr></mj-table>"
	case 8:
		var b strings.Builder
		b.WriteString("<mj-social" + f.attrs("mj-social", 3) + ">")
		for range 1 + f.r.IntN(3) {
			b.WriteString(`<mj-social-element name="` + f.pick([]string{"facebook", "twitter", "linkedin", "github", "custom"}) + `"` + f.attrs("mj-social-element", 2, "name") + ">" + f.label() + "</mj-social-element>")
		}
		return b.String() + "</mj-social>"
	case 9:
		var b strings.Builder
		b.WriteString("<mj-navbar" + f.attrs("mj-navbar", 2) + ">")
		for range 1 + f.r.IntN(3) {
			b.WriteString(`<mj-navbar-link href="/p"` + f.attrs("mj-navbar-link", 2, "href") + ">" + f.label() + "</mj-navbar-link>")
		}
		return b.String() + "</mj-navbar>"
	case 10:
		var b strings.Builder
		b.WriteString("<mj-accordion" + f.attrs("mj-accordion", 2) + ">")
		for range 1 + f.r.IntN(2) {
			b.WriteString("<mj-accordion-element" + f.attrs("mj-accordion-element", 1) + "><mj-accordion-title" + f.attrs("mj-accordion-title", 1) + ">" + f.label() + "</mj-accordion-title><mj-accordion-text" + f.attrs("mj-accordion-text", 1) + ">" + f.text() + "</mj-accordion-text></mj-accordion-element>")
		}
		return b.String() + "</mj-accordion>"
	default:
		var b strings.Builder
		b.WriteString("<mj-carousel" + f.attrs("mj-carousel", 2) + ">")
		for i := range 2 + f.r.IntN(2) {
			fmt.Fprintf(&b, `<mj-carousel-image src="https://example.com/%d.png"%s />`, i, f.attrs("mj-carousel-image", 1, "src"))
		}
		return b.String() + "</mj-carousel>"
	}
}

// edgeDocuments are pathological inputs TestDifferentialCrashes renders with gomjml alone: sizes,
// depths and values no email would have, where a crash or superlinear time would hide.
func edgeDocuments() map[string]string {
	return map[string]string{
		"2000 sections":        mjmlDocument("", "", strings.Repeat(inColumn("<mj-text>x</mj-text>"), 2000)),
		"500 columns":          mjmlDocument("", "", "<mj-section>"+strings.Repeat("<mj-column><mj-text>x</mj-text></mj-column>", 500)+"</mj-section>"),
		"200 nested wrappers":  mjmlDocument("", "", strings.Repeat("<mj-wrapper>", 200)+inColumn("<mj-text>x</mj-text>")+strings.Repeat("</mj-wrapper>", 200)),
		"3000 nested divs":     mjmlDocument("", "", inColumn("<mj-text>"+strings.Repeat("<div>", 3000)+"x"+strings.Repeat("</div>", 3000)+"</mj-text>")),
		"5000 nested spans":    mjmlDocument("", "", "<mj-raw>"+strings.Repeat("<span>", 5000)+strings.Repeat("</span>", 5000)+"</mj-raw>"),
		"200 kB attribute":     mjmlDocument("", "", inColumn(`<mj-text css-class="`+strings.Repeat("a ", 100000)+`">x</mj-text>`)),
		"overflowing lengths":  mjmlDocument("", ` width="99999999999999999999px"`, inColumn(`<mj-image width="999999999999px" src="x" /><mj-text padding="99999999999999999999px">x</mj-text>`)),
		"zero widths":          mjmlDocument("", "", `<mj-section><mj-column width="0%"><mj-image src="x"/></mj-column><mj-column width="0px"><mj-text>x</mj-text></mj-column></mj-section>`),
		"negative lengths":     mjmlDocument("", "", `<mj-section padding="-10px"><mj-column width="-50%" padding="-5px"><mj-image width="-20px" src="x"/><mj-divider width="-10%"/></mj-column></mj-section>`),
		"non-numeric lengths":  mjmlDocument("", ` width="NaNpx"`, `<mj-section><mj-column width="abc%"><mj-image width="1e308px" src="x"/></mj-column></mj-section>`),
		"empty everything":     "<mjml><mj-head></mj-head><mj-body><mj-section><mj-column></mj-column></mj-section><mj-wrapper></mj-wrapper><mj-hero></mj-hero><mj-raw></mj-raw></mj-body></mjml>",
		"empty containers":     mjmlDocument("", "", inColumn("<mj-social></mj-social><mj-navbar></mj-navbar><mj-accordion></mj-accordion><mj-carousel></mj-carousel><mj-table></mj-table>")),
		"components misplaced": "<mjml><mj-body><mj-text>top</mj-text><mj-column><mj-section><mj-button>x</mj-button></mj-section></mj-column></mj-body></mjml>",
	}
}

var mutationTokens = []string{"<", ">", "</mj-column>", "</mj-section>", "<mj-section>", "<mj-column>", `"`, "&", "&#", "<!--",
	"-->", "<![CDATA[", "<mj-text>", "</mj-text>", "<mj-raw>", "<mj-include path=\"x.mjml\" />", "=", " mj-class=\"c0\"", "\x00", "<mj-body>", "</mjml>"}

// mutate damages a document for crash testing, driven by the bytes of data.
func mutate(document string, data []byte) string {
	for i := 0; i+2 < len(data) && len(document) > 0; i += 3 {
		at := int(data[i+1]) * len(document) / 256
		switch data[i] % 5 {
		case 0: // delete a range
			end := min(at+int(data[i+2])%32+1, len(document))
			document = document[:at] + document[end:]
		case 1: // insert a token
			document = document[:at] + mutationTokens[int(data[i+2])%len(mutationTokens)] + document[at:]
		case 2: // duplicate a range
			end := min(at+int(data[i+2])%64+1, len(document))
			document = document[:end] + document[at:end] + document[end:]
		case 3: // truncate
			document = document[:at]
		case 4: // swap two halves around the point
			document = document[at:] + document[:at]
		}
	}
	return document
}
