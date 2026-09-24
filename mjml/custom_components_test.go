package mjml

import (
	"encoding/xml"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/custom"
)

func mustRegister(t *testing.T, reg *custom.Registry, tag string, def custom.Def) {
	t.Helper()
	if err := reg.Register(tag, def); err != nil {
		t.Fatal(err)
	}
}

func mustRender(t *testing.T, src string, opts ...RenderOption) string {
	t.Helper()
	out, err := Render(src, opts...)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func policySummary(t *testing.T) *custom.Registry {
	t.Helper()
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-policy-summary", custom.Def{
		Attributes: []string{"insured", "premium", "currency"},
		Defaults:   map[string]string{"currency": "GBP"},
		Expand: func(c custom.Call) ([]*custom.Node, error) {
			return []*custom.Node{
				custom.Element("mj-text", map[string]string{"font-size": "18px"},
					custom.Text("Policy for "+c.Attrs["insured"])),
				custom.Element("mj-table", nil,
					custom.Element("tr", nil,
						custom.Element("td", nil, custom.Text("Premium")),
						custom.Element("td", nil, custom.Text(c.Attrs["currency"]+" "+c.Attrs["premium"])))),
			}, nil
		},
	})
	return reg
}

func inColumn(content string) string {
	return `<mjml><mj-body><mj-section><mj-column>` + content + `</mj-column></mj-section></mj-body></mjml>`
}

func TestCustomComponentRendersAsItsExpansion(t *testing.T) {
	got := mustRender(t,
		inColumn(`<mj-policy-summary insured="O'Brien &amp; Sons" premium="1,234.00" />`),
		WithComponents(policySummary(t)))
	want := mustRender(t, inColumn(
		`<mj-text font-size="18px">Policy for O&#39;Brien &amp; Sons</mj-text>`+
			`<mj-table><tr><td>Premium</td><td>GBP 1,234.00</td></tr></mj-table>`))
	if got != want {
		t.Fatalf("custom tag differs from its hand-written expansion\n got: %s\nwant: %s", got, want)
	}
}

func TestCustomComponentAttributesFollowMJMLPrecedence(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-probe", custom.Def{
		Defaults: map[string]string{"a": "default", "b": "default", "c": "default", "d": "default"},
		Expand: func(c custom.Call) ([]*custom.Node, error) {
			return custom.Parse(fmt.Sprintf(`<mj-text>a=%s b=%s c=%s d=%s font=%q</mj-text>`,
				c.Attrs["a"], c.Attrs["b"], c.Attrs["c"], c.Attrs["d"], c.Attrs["font-family"]))
		},
	})
	src := `<mjml><mj-head><mj-attributes>
<mj-all font-family="Arial" />
<mj-class name="k" a="class" b="class" />
<mj-probe a="tag" b="tag" c="tag" />
</mj-attributes></mj-head>
<mj-body><mj-section><mj-column><mj-probe mj-class="k" a="element" /></mj-column></mj-section></mj-body></mjml>`

	got := mustRender(t, src, WithComponents(reg))
	if !strings.Contains(got, `a=element b=class c=tag d=default font=""`) {
		t.Fatalf("unexpected precedence in %s", got)
	}
}

func TestCustomComponentReportsUndeclaredAttributesAtItsLine(t *testing.T) {
	src := "<mjml><mj-body><mj-section><mj-column>\n\n" +
		`<mj-policy-summary insured="Acme" colour="red" custom-note="x" data-id="1" />` +
		"\n</mj-column></mj-section></mj-body></mjml>"

	out, err := Render(src, WithComponents(policySummary(t)))
	var mjmlErr Error
	if !errors.As(err, &mjmlErr) || len(mjmlErr.Details) != 2 {
		t.Fatalf("want colour and custom-note reported, got %v", err)
	}
	if d := mjmlErr.Details[0]; d.TagName != "mj-policy-summary" || d.Line != 3 || !strings.Contains(d.Message, "colour") {
		t.Fatalf("unexpected detail %+v", d)
	}
	if !strings.Contains(out, "Policy for Acme") {
		t.Fatal("the HTML should still be returned alongside the validation error")
	}

	_, err = Render(src, WithComponents(policySummary(t)), WithAllowedAttributes(func(tagName, attrName string) bool {
		return tagName == "mj-policy-summary" && strings.HasPrefix(attrName, "custom-")
	}))
	if !errors.As(err, &mjmlErr) || len(mjmlErr.Details) != 1 || !strings.Contains(mjmlErr.Details[0].Message, "colour") {
		t.Fatalf("WithAllowedAttributes should accept custom-note on a custom tag, got %v", err)
	}
}

func TestCustomComponentAcceptsAttributesWithDefaults(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-tone", custom.Def{
		Attributes: []string{},
		Defaults:   map[string]string{"tone": "calm"},
		Expand: func(c custom.Call) ([]*custom.Node, error) {
			return custom.Parse(`<mj-text>` + c.Attrs["tone"] + `</mj-text>`)
		},
	})
	_, err := Render(inColumn(`<mj-tone tone="loud" volume="11" />`), WithComponents(reg))
	var mjmlErr Error
	if !errors.As(err, &mjmlErr) || len(mjmlErr.Details) != 1 || !strings.Contains(mjmlErr.Details[0].Message, "volume") {
		t.Fatalf("want only volume reported, got %v", err)
	}
}

func TestCustomComponentRejectsANilNode(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-nil", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) { return []*custom.Node{nil}, nil },
	})
	if _, err := Render(inColumn(`<mj-nil />`), WithComponents(reg)); err == nil ||
		!strings.Contains(err.Error(), "<mj-nil>: expanded to a nil node") {
		t.Fatalf("got %v", err)
	}
}

func TestCustomComponentReportsInvalidAttributesInItsExpansion(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-broken", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) {
			return custom.Parse(`<mj-text bogus="1">x</mj-text>`)
		},
	})
	_, err := Render(inColumn("\n\n<mj-broken a=\"1\" />"), WithComponents(reg))
	var mjmlErr Error
	if !errors.As(err, &mjmlErr) || len(mjmlErr.Details) != 1 {
		t.Fatalf("want one detail, got %v", err)
	}
	if d := mjmlErr.Details[0]; d.TagName != "mj-text" || d.Line != 3 {
		t.Fatalf("an error inside the expansion should point at the custom tag, got %+v", d)
	}
}

func TestCustomComponentContainerExpandsItsChildren(t *testing.T) {
	reg := policySummary(t)
	mustRegister(t, reg, "mj-policy-card", custom.Def{
		Expand: func(c custom.Call) ([]*custom.Node, error) {
			if c.Parent != "mj-body" {
				return nil, fmt.Errorf("must be placed in mj-body, not %s", c.Parent)
			}
			section, err := custom.Parse(`<mj-section background-color="#eeeeee"><mj-column></mj-column></mj-section>`)
			if err != nil {
				return nil, err
			}
			section[0].Children[0].Children = c.Node.Children
			return section, nil
		},
	})

	got := mustRender(t, `<mjml><mj-body><mj-policy-card>`+
		`<mj-policy-summary insured="Acme" premium="10" /><mj-text>Thanks</mj-text>`+
		`</mj-policy-card></mj-body></mjml>`, WithComponents(reg))
	want := mustRender(t, `<mjml><mj-body><mj-section background-color="#eeeeee"><mj-column>`+
		`<mj-text font-size="18px">Policy for Acme</mj-text>`+
		`<mj-table><tr><td>Premium</td><td>GBP 10</td></tr></mj-table><mj-text>Thanks</mj-text>`+
		`</mj-column></mj-section></mj-body></mjml>`)
	if got != want {
		t.Fatalf("container differs from its hand-written expansion\n got: %s\nwant: %s", got, want)
	}

	_, err := Render(inColumn(`<mj-policy-card />`), WithComponents(reg))
	if err == nil || !strings.Contains(err.Error(), "must be placed in mj-body, not mj-column") {
		t.Fatalf("want the expander's error, got %v", err)
	}
}

func TestCustomComponentErrorCarriesItsLine(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-fail", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) { return nil, errors.New("no quote") },
	})
	_, err := Render(inColumn("\n\n<mj-fail a=\"1\" />"), WithComponents(reg))
	if err == nil || !strings.Contains(err.Error(), "line 3: <mj-fail>: no quote") {
		t.Fatalf("got %v", err)
	}
}

func TestCustomComponentRecursionIsBounded(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-loop", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) { return custom.Parse(`<mj-loop />`) },
	})
	_, err := Render(inColumn(`<mj-loop />`), WithComponents(reg))
	if err == nil || !strings.Contains(err.Error(), "<mj-loop> expands deeper than") {
		t.Fatalf("got %v", err)
	}
}

func TestCustomComponentHeadStyleIsAddedOnceBeforeTheDocumentsOwn(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-note", custom.Def{
		HeadStyle: ".note { color: red; }",
		Expand: func(custom.Call) ([]*custom.Node, error) {
			return custom.Parse(`<mj-text css-class="note">n</mj-text>`)
		},
	})
	body := `<mj-body><mj-section><mj-column>%s%s</mj-column></mj-section></mj-body></mjml>`
	note := `<mj-text css-class="note">n</mj-text>`
	style := `<mj-style>.note { color: red; }</mj-style>`
	mine := `<mj-style>.mine { color: blue; }</mj-style>`

	for name, tc := range map[string]struct{ head, want string }{
		"without mj-head": {"", `<mj-head>` + style + `</mj-head>`},
		"with mj-head":    {`<mj-head>` + mine + `</mj-head>`, `<mj-head>` + style + mine + `</mj-head>`},
	} {
		t.Run(name, func(t *testing.T) {
			got := mustRender(t, `<mjml>`+tc.head+fmt.Sprintf(body, `<mj-note />`, `<mj-note />`), WithComponents(reg))
			want := mustRender(t, `<mjml>`+tc.want+fmt.Sprintf(body, note, note))
			if got != want {
				t.Fatalf("head style differs from its hand-written equivalent\n got: %s\nwant: %s", got, want)
			}
		})
	}
}

func TestCustomComponentLeavesTheASTUnexpanded(t *testing.T) {
	src := inColumn(`<mj-policy-summary insured="Acme" />`)
	reg := policySummary(t)

	var wg sync.WaitGroup
	for i := range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			opts := []RenderOption{WithCache()}
			if i%2 == 0 {
				opts = append(opts, WithComponents(reg))
			}
			res, err := RenderWithAST(src, opts...)
			if err != nil {
				t.Error(err)
				return
			}
			if strings.Contains(res.HTML, "Policy for Acme") != (i%2 == 0) {
				t.Errorf("render %d: a registry leaked between renders", i)
			}
			column := res.AST.FindFirstChild("mj-body").Children[0].Children[0]
			if column.Children[0].GetTagName() != "mj-policy-summary" {
				t.Errorf("render %d: the returned AST should keep the custom tag", i)
			}
		}()
	}
	wg.Wait()
}

func TestCustomComponentExpandsFromAST(t *testing.T) {
	ast, err := ParseMJML(inColumn(`<mj-policy-summary insured="Acme" />`))
	if err != nil {
		t.Fatal(err)
	}
	got, err := RenderFromAST(ast, WithComponents(policySummary(t)))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "Policy for Acme") {
		t.Fatal("RenderFromAST should expand custom tags")
	}
}

func TestCustomComponentCannotRedefineABuiltinTag(t *testing.T) {
	file, err := parser.ParseFile(token.NewFileSet(), "component.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var tags []string
	ast.Inspect(file, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Name.Name != "CreateComponent" {
			return false
		}
		if clause, ok := n.(*ast.CaseClause); ok {
			for _, expr := range clause.List {
				if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					tag, _ := strconv.Unquote(lit.Value)
					tags = append(tags, tag)
				}
			}
		}
		return true
	})
	if len(tags) < 30 {
		t.Fatalf("found only %d tags in CreateComponent", len(tags))
	}

	reg := custom.NewRegistry()
	expand := func(custom.Call) ([]*custom.Node, error) { return nil, nil }
	for _, tag := range tags {
		if err := reg.Register(tag, custom.Def{Expand: expand}); err == nil {
			t.Errorf("registering built-in <%s> should fail", tag)
		}
	}
}

func TestCustomComponentMayShareTheMJTextPrefix(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-text-block", custom.Def{
		Expand: func(c custom.Call) ([]*custom.Node, error) { return c.Node.Children, nil },
	})
	got := mustRender(t, inColumn(`<mj-text-block><mj-image src="https://x/y.png" /></mj-text-block><mj-text>hi</mj-text>`),
		WithComponents(reg))
	want := mustRender(t, inColumn(`<mj-image src="https://x/y.png" /><mj-text>hi</mj-text>`))
	if got != want {
		t.Fatalf("mj-text-block differs from its expansion\n got: %s\nwant: %s", got, want)
	}
}

func TestCustomComponentErrorLinesWithoutAttributes(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-broken", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) { return custom.Parse(`<mj-text bogus="1">x</mj-text>`) },
	})
	mustRegister(t, reg, "mj-fail", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) { return nil, errors.New("no quote") },
	})

	_, err := Render(inColumn("\n\n<mj-broken />"), WithComponents(reg))
	var mjmlErr Error
	if !errors.As(err, &mjmlErr) || len(mjmlErr.Details) != 1 || mjmlErr.Details[0].Line != 3 {
		t.Fatalf("want bogus reported at line 3, got %v", err)
	}
	_, err = Render(inColumn("\n\n<mj-fail />"), WithComponents(reg))
	if err == nil || !strings.Contains(err.Error(), "line 3: <mj-fail>: no quote") {
		t.Fatalf("got %v", err)
	}
}

func TestCustomComponentCallCannotWriteIntoTheCachedAST(t *testing.T) {
	var shared atomic.Bool
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-card", custom.Def{
		Expand: func(c custom.Call) ([]*custom.Node, error) {
			n := c.Node
			if cap(n.Children) != len(n.Children) || cap(n.MixedContent) != len(n.MixedContent) || cap(n.Attrs) != len(n.Attrs) {
				shared.Store(true)
			}
			foot, err := custom.Parse(`<mj-text>foot</mj-text>`)
			if err != nil {
				return nil, err
			}
			section, err := custom.Parse(`<mj-section><mj-column></mj-column></mj-section>`)
			if err != nil {
				return nil, err
			}
			section[0].Children[0].Children = append(n.Children, foot...)
			n.Attrs = append(n.Attrs, xml.Attr{Name: xml.Name{Local: "seen"}})
			return section, nil
		},
	})
	src := `<mjml><mj-body><mj-card title="t"><mj-text>a</mj-text><mj-text>b</mj-text><mj-text>c</mj-text></mj-card></mj-body></mjml>`

	var wg sync.WaitGroup
	for range 64 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := Render(src, WithCache(), WithComponents(reg)); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if shared.Load() {
		t.Fatal("an append to Call.Node's slices would write into the cached AST")
	}
}

func TestCustomComponentRegistryIsSafeForConcurrentUse(t *testing.T) {
	reg := policySummary(t)
	src := inColumn(`<mj-policy-summary insured="Acme" />`)
	expand := func(custom.Call) ([]*custom.Node, error) { return nil, nil }

	var wg sync.WaitGroup
	for i := range 32 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			if _, err := Render(src, WithComponents(reg)); err != nil {
				t.Error(err)
			}
		}()
		go func() {
			defer wg.Done()
			mustRegister(t, reg, fmt.Sprintf("mj-other-%d", i), custom.Def{Expand: expand})
		}()
	}
	wg.Wait()
}

func TestCustomComponentTextCannotBecomeMarkup(t *testing.T) {
	const evil = `<img src=x onerror=alert(1)>`
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-evil", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) {
			return []*custom.Node{
				custom.Element("mj-text", nil, custom.Text(evil)),
				custom.Element("mj-text", nil, custom.Element("b", nil, custom.Text(evil))),
				custom.Element("mj-table", nil,
					custom.Element("tr", nil, custom.Element("td", nil, custom.Text(evil)))),
			}, nil
		},
	})
	got := mustRender(t, inColumn(`<mj-evil />`), WithComponents(reg))
	if strings.Contains(got, "<img") {
		t.Fatalf("text became live markup: %s", got)
	}
	if n := strings.Count(got, "&lt;img src=x onerror=alert(1)&gt;"); n != 3 {
		t.Fatalf("want the text shown literally 3 times, got %d", n)
	}
}

func TestCustomComponentRejectsTextOutsideAnElement(t *testing.T) {
	reg := custom.NewRegistry()
	mustRegister(t, reg, "mj-bare", custom.Def{
		Expand: func(custom.Call) ([]*custom.Node, error) { return []*custom.Node{custom.Text("x")}, nil },
	})
	if _, err := Render(inColumn(`<mj-bare />`), WithComponents(reg)); err == nil ||
		!strings.Contains(err.Error(), "text outside an element") {
		t.Fatalf("got %v", err)
	}
}
