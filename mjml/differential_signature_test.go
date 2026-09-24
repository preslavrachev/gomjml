package mjml

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"golang.org/x/net/html"
)

// differenceSignatures describes how gomjml's output differs from the reference as a set of
// normalised signatures, "<component> <kind> <target>", such as
// "mj-button style-changed td{padding}". Values are left out, so that one defect gives the same
// signature in every case it shows up in, and cases can be clustered by it.
//
// The component is the one gomjml's debug attributes name on the element or its nearest
// ancestor, so debug is gomjml's output rendered WithDebugTags. It is only used to name things:
// whether two outputs are equivalent is decided by compareWithReference.
func differenceSignatures(reference, debug string) []string {
	refDoc, err1 := html.Parse(strings.NewReader(normalizeForComparison(reference)))
	gotDoc, err2 := html.Parse(strings.NewReader(normalizeForComparison(debug)))
	if err1 != nil || err2 != nil {
		return []string{"document unparsable"}
	}
	s := &signer{set: map[string]bool{}}
	s.children(refDoc, gotDoc, "document")
	s.css(styleSheets(refDoc), styleSheets(gotDoc))
	return slices.Sorted(maps.Keys(s.set))
}

type signer struct{ set map[string]bool }

func (s *signer) add(component, kind, target string) {
	s.set[component+" "+kind+" "+target] = true
}

// componentOf names the component whose debug attribute an element carries, if any.
func componentOf(node *html.Node, inherited string) string {
	if node == nil || node.Type != html.ElementNode {
		return inherited
	}
	switch node.Data {
	case "head":
		return "head"
	case "body":
		return "mj-body"
	}
	for _, attr := range node.Attr {
		if name, ok := strings.CutPrefix(attr.Key, "data-mj-debug-"); ok {
			return "mj-" + name
		}
	}
	return inherited
}

// item is a child the alignment works on: an element, a run of text or a conditional comment.
type item struct {
	key  string
	node *html.Node
	text string
	raw  string // a text run as written, trimmed
}

func items(parent *html.Node) []item {
	var list []item
	var text strings.Builder
	flush := func() {
		if run := strings.Join(strings.Fields(text.String()), " "); run != "" {
			list = append(list, item{key: "#text", text: run, raw: strings.TrimSpace(text.String())})
		}
		text.Reset()
	}
	for child := parent.FirstChild; child != nil; child = child.NextSibling {
		switch child.Type {
		case html.TextNode:
			text.WriteString(child.Data)
		case html.ElementNode:
			if child.Data == "style" {
				continue // compared as CSS
			}
			flush()
			list = append(list, item{key: child.Data, node: child})
		case html.CommentNode:
			if match := conditionalComment.FindStringSubmatch(child.Data); match != nil {
				flush()
				list = append(list, item{key: "#mso " + match[1] + firstTag(match[2]), node: child, text: child.Data})
			}
		}
	}
	flush()
	return list
}

var firstTagPattern = regexp.MustCompile(`<(/?[a-zA-Z:]+)`)

func firstTag(markup string) string {
	if match := firstTagPattern.FindStringSubmatch(markup); match != nil {
		return " " + match[1]
	}
	return ""
}

// align pairs the items of two lists by key with a longest common subsequence.
func align[T any](a, b []T, key func(T) string) (pairs [][2]int) {
	n, m := len(a), len(b)
	lcs := make([][]int, n+1)
	for i := range lcs {
		lcs[i] = make([]int, m+1)
	}
	for i := n - 1; i >= 0; i-- {
		for j := m - 1; j >= 0; j-- {
			if key(a[i]) == key(b[j]) {
				lcs[i][j] = lcs[i+1][j+1] + 1
			} else {
				lcs[i][j] = max(lcs[i+1][j], lcs[i][j+1])
			}
		}
	}
	i, j := 0, 0
	for i < n || j < m {
		switch {
		case i < n && j < m && key(a[i]) == key(b[j]):
			pairs = append(pairs, [2]int{i, j})
			i++
			j++
		case j == m || (i < n && lcs[i+1][j] >= lcs[i][j+1]):
			pairs = append(pairs, [2]int{i, -1})
			i++
		default:
			pairs = append(pairs, [2]int{-1, j})
			j++
		}
	}
	return pairs
}

// msoComponent names the component an MSO block at gotItems[j] belongs to: gomjml writes a
// component's opening conditional just before its element and the closing one just after.
func msoComponent(gotItems []item, j int, inherited string) string {
	for k := j + 1; k < len(gotItems); k++ {
		if gotItems[k].node != nil && gotItems[k].node.Type == html.ElementNode {
			if component := componentOf(gotItems[k].node, inherited); component != inherited {
				return component
			}
			break
		}
	}
	for k := j - 1; k >= 0; k-- {
		if gotItems[k].node != nil && gotItems[k].node.Type == html.ElementNode {
			return componentOf(gotItems[k].node, inherited)
		}
	}
	return inherited
}

func (s *signer) children(ref, got *html.Node, component string) {
	refItems, gotItems := items(ref), items(got)
	for _, pair := range align(refItems, gotItems, func(it item) string { return it.key }) {
		switch {
		case pair[1] == -1:
			r := refItems[pair[0]]
			s.add(component, "missing", describeItem(r)+" in "+parentName(ref))
		case pair[0] == -1:
			g := gotItems[pair[1]]
			extra := componentOf(g.node, component)
			if strings.HasPrefix(g.key, "#mso") {
				extra = msoComponent(gotItems, pair[1], component)
			}
			s.add(extra, "extra", describeItem(g)+" in "+parentName(got))
		default:
			r, g := refItems[pair[0]], gotItems[pair[1]]
			switch {
			case r.key == "#text":
				switch {
				case r.text != g.text:
					s.add(component, "text-changed", "text in "+parentName(ref))
				case r.raw != g.raw:
					// compareNodes trims text but keeps the whitespace inside it.
					s.add(component, "whitespace-changed", "text in "+parentName(ref))
				}
			case strings.HasPrefix(r.key, "#mso"):
				s.mso(r.text, g.text, msoComponent(gotItems, pair[1], component))
			default:
				inner := componentOf(g.node, component)
				s.attributes("", g.node.Data, r.node.Attr, g.node.Attr, inner)
				if g.node.Data == "body" {
					// Only mj-body's own attributes are its; below, no debug attribute names a component.
					inner = "body"
				}
				s.children(r.node, g.node, inner)
			}
		}
	}
}

func parentName(node *html.Node) string {
	if node.Type == html.DocumentNode {
		return "document"
	}
	return node.Data
}

func describeItem(it item) string {
	switch {
	case it.key == "#text":
		return "text"
	case strings.HasPrefix(it.key, "#mso"):
		return "mso-block" + strings.TrimPrefix(it.key, "#mso")
	}
	return "element " + it.key
}

// attributes compares the attributes of two matched tags. prefix marks tags inside MSO blocks.
func (s *signer) attributes(prefix, tag string, ref, got []html.Attribute, component string) {
	refAttrs, gotAttrs := attributeMap(ref), attributeMap(got)
	for _, key := range slices.Sorted(mapsKeysUnion(refAttrs, gotAttrs)) {
		r, inRef := refAttrs[key]
		g, inGot := gotAttrs[key]
		switch {
		case key == "style":
			s.style(prefix, tag, r, g, component)
		case key == "class":
			s.classes(prefix, tag, r, g, component)
		case !inGot:
			s.add(component, prefix+"attr-missing", fmt.Sprintf("%s[%s]", tag, key))
		case !inRef:
			s.add(component, prefix+"attr-extra", fmt.Sprintf("%s[%s]", tag, key))
		case r != g:
			s.add(component, prefix+"attr-changed", fmt.Sprintf("%s[%s]", tag, key))
		}
	}
}

func attributeMap(attrs []html.Attribute) map[string]string {
	m := make(map[string]string, len(attrs))
	for _, attr := range attrs {
		if !strings.HasPrefix(attr.Key, "data-mj-debug") {
			m[attr.Key] = attr.Val
		}
	}
	return m
}

func mapsKeysUnion[V any](a, b map[string]V) func(func(string) bool) {
	return func(yield func(string) bool) {
		for k := range a {
			if !yield(k) {
				return
			}
		}
		for k := range b {
			if _, dup := a[k]; !dup && !yield(k) {
				return
			}
		}
	}
}

// valueSpace finds whitespace that cannot change a declaration's value.
var valueSpace = regexp.MustCompile(`\s*([!,()/])\s*`)

type declaration struct{ property, value string }

func declarations(style string) []declaration {
	var list []declaration
	for part := range strings.SplitSeq(style, ";") {
		if property, value, found := strings.Cut(part, ":"); found {
			value = valueSpace.ReplaceAllString(strings.Join(strings.Fields(value), " "), "$1")
			list = append(list, declaration{strings.ToLower(strings.TrimSpace(property)), value})
		}
	}
	return list
}

func (s *signer) style(prefix, tag, ref, got, component string) {
	refDecls, gotDecls := declarations(ref), declarations(got)
	effective := func(list []declaration) map[string]string {
		m := map[string]string{}
		for _, d := range list {
			m[d.property] = d.value
		}
		return m
	}
	refMap, gotMap := effective(refDecls), effective(gotDecls)
	for _, property := range slices.Sorted(mapsKeysUnion(refMap, gotMap)) {
		r, inRef := refMap[property]
		g, inGot := gotMap[property]
		target := fmt.Sprintf("%s{%s}", tag, property)
		switch {
		case !inGot:
			s.add(component, prefix+"style-missing", target)
		case !inRef:
			s.add(component, prefix+"style-extra", target)
		case r != g:
			s.add(component, prefix+"style-changed", target)
		}
	}
	names := func(list []declaration) []string {
		var out []string
		for _, d := range list {
			out = append(out, d.property)
		}
		return out
	}
	refNames, gotNames := names(refDecls), names(gotDecls)
	for _, list := range [][]string{refNames, gotNames} {
		if duplicate := firstDuplicate(list); duplicate != "" && firstDuplicate(refNames) != firstDuplicate(gotNames) {
			s.add(component, prefix+"style-duplicate", fmt.Sprintf("%s{%s}", tag, duplicate))
		}
	}
	for i, first := range refNames {
		for _, second := range refNames[i+1:] {
			if !resetsLonghand(first, second) && !resetsLonghand(second, first) {
				continue
			}
			a, b := slices.Index(gotNames, first), slices.Index(gotNames, second)
			if a != -1 && b != -1 && a > b {
				s.add(component, prefix+"style-order", fmt.Sprintf("%s{%s,%s}", tag, first, second))
			}
		}
	}
}

var digits = regexp.MustCompile(`\d+`)

func (s *signer) classes(prefix, tag, ref, got, component string) {
	refSet, gotSet := strings.Fields(ref), strings.Fields(got)
	for _, class := range refSet {
		if !slices.Contains(gotSet, class) {
			s.add(component, prefix+"class-missing", tag+"."+digits.ReplaceAllString(class, "N"))
		}
	}
	for _, class := range gotSet {
		if !slices.Contains(refSet, class) {
			s.add(component, prefix+"class-extra", tag+"."+digits.ReplaceAllString(class, "N"))
		}
	}
}

// msoToken is a tag or text run inside a conditional comment.
type msoToken struct {
	key   string // "<tag", "</tag" or "#text"
	tag   string
	attrs []html.Attribute
	text  string
}

func msoTokens(markup string) []msoToken {
	var tokens []msoToken
	tokenizer := html.NewTokenizer(strings.NewReader(markup))
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			return tokens
		case html.StartTagToken, html.SelfClosingTagToken:
			token := tokenizer.Token()
			tokens = append(tokens, msoToken{key: "<" + token.Data, tag: token.Data, attrs: token.Attr})
		case html.EndTagToken:
			token := tokenizer.Token()
			tokens = append(tokens, msoToken{key: "</" + token.Data, tag: token.Data})
		case html.TextToken:
			if text := strings.Join(strings.Fields(string(tokenizer.Text())), " "); text != "" {
				tokens = append(tokens, msoToken{key: "#text", text: text})
			}
		}
	}
}

func (s *signer) mso(ref, got, component string) {
	refMatch, gotMatch := conditionalComment.FindStringSubmatch(ref), conditionalComment.FindStringSubmatch(got)
	if refMatch[1] != gotMatch[1] || (refMatch[3] == "") != (gotMatch[3] == "") {
		s.add(component, "mso-condition-changed", refMatch[1])
	}
	refTokens, gotTokens := msoTokens(refMatch[2]), msoTokens(gotMatch[2])
	for _, pair := range align(refTokens, gotTokens, func(t msoToken) string { return t.key }) {
		switch {
		case pair[1] == -1:
			s.add(component, "mso-missing", describeMSOToken(refTokens[pair[0]]))
		case pair[0] == -1:
			s.add(component, "mso-extra", describeMSOToken(gotTokens[pair[1]]))
		default:
			r, g := refTokens[pair[0]], gotTokens[pair[1]]
			switch {
			case r.key == "#text" && r.text != g.text:
				s.add(component, "mso-text-changed", "text")
			case strings.HasPrefix(r.key, "<"):
				s.attributes("mso-", r.tag, r.attrs, g.attrs, component)
			}
		}
	}
}

func describeMSOToken(t msoToken) string {
	if t.key == "#text" {
		return "text"
	}
	return "tag " + t.key + ">"
}

// styleSheets returns the text of every <style> element outside conditional comments, in order.
func styleSheets(doc *html.Node) []string {
	var sheets []string
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "style" {
			var text strings.Builder
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				text.WriteString(child.Data)
			}
			sheets = append(sheets, text.String())
			return
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(doc)
	return sheets
}

type cssDeclaration struct{ context, selector, property, value string }

var cssComment = regexp.MustCompile(`(?s)/\*.*?\*/`)

// cssDeclarations flattens style sheets into declarations with their at-rule and selector.
func cssDeclarations(sheets []string) []cssDeclaration {
	var list []cssDeclaration
	for _, sheet := range sheets {
		css := cssComment.ReplaceAllString(sheet, "")
		var context []string
		var buffer strings.Builder
		selector := ""
		for _, r := range css {
			switch r {
			case '{':
				prelude := strings.Join(strings.Fields(buffer.String()), " ")
				buffer.Reset()
				if strings.HasPrefix(prelude, "@") && !strings.HasPrefix(prelude, "@font-face") {
					context = append(context, prelude)
				} else {
					selector = prelude
				}
			case '}':
				if selector != "" {
					for _, d := range declarations(buffer.String()) {
						list = append(list, cssDeclaration{strings.Join(context, " "), selector, d.property, d.value})
					}
					selector = ""
				} else if len(context) > 0 {
					context = context[:len(context)-1]
				}
				buffer.Reset()
			case ';':
				if selector == "" {
					if statement := strings.Join(strings.Fields(buffer.String()), " "); strings.HasPrefix(statement, "@") {
						list = append(list, cssDeclaration{strings.Join(context, " "), statement, "", ""})
					}
					buffer.Reset()
				} else {
					buffer.WriteRune(r)
				}
			default:
				buffer.WriteRune(r)
			}
		}
	}
	return list
}

// combinatorSpace finds the optional whitespace around selector combinators and commas.
var combinatorSpace = regexp.MustCompile(`\s*([+>~,:()])\s*`)

// cssComponents names the component a head rule belongs to by its selector.
var cssComponents = []struct{ fragment, component string }{
	{"mj-column", "mj-column"}, {"mj-carousel", "mj-carousel"}, {"mj-menu", "mj-navbar"},
	{"mj-inline-links", "mj-navbar"}, {"hamburger", "mj-navbar"}, {"mj-accordion", "mj-accordion"},
	{"mj-full-width-mobile", "mj-image"}, {"mj-social", "mj-social"}, {"mj-group", "mj-group"},
}

func cssComponent(selector string) string {
	for _, c := range cssComponents {
		if strings.Contains(selector, c.fragment) {
			return c.component
		}
	}
	return "head"
}

func (s *signer) css(refSheets, gotSheets []string) {
	refDecls, gotDecls := cssDeclarations(refSheets), cssDeclarations(gotSheets)
	key := func(d cssDeclaration) string {
		return combinatorSpace.ReplaceAllString(d.context+"|"+d.selector, "$1") + "|" + d.property
	}
	shape := func(d cssDeclaration) string {
		target := digits.ReplaceAllString(combinatorSpace.ReplaceAllString(d.selector, "$1"), "N")
		if d.property != "" {
			target += "{" + d.property + "}"
		}
		if d.context != "" {
			target = digits.ReplaceAllString(d.context, "N") + " " + target
		}
		return target
	}
	refByKey := map[string][]cssDeclaration{}
	for _, d := range refDecls {
		refByKey[key(d)] = append(refByKey[key(d)], d)
	}
	gotByKey := map[string][]cssDeclaration{}
	for _, d := range gotDecls {
		gotByKey[key(d)] = append(gotByKey[key(d)], d)
	}
	for _, k := range slices.Sorted(mapsKeysUnion(refByKey, gotByKey)) {
		r, g := refByKey[k], gotByKey[k]
		switch {
		case len(g) == 0:
			s.add(cssComponent(r[0].selector), "css-missing", shape(r[0]))
		case len(r) == 0:
			s.add(cssComponent(g[0].selector), "css-extra", shape(g[0]))
		case r[len(r)-1].value != g[len(g)-1].value:
			s.add(cssComponent(r[0].selector), "css-changed", shape(r[0]))
		case len(r) != len(g):
			s.add(cssComponent(r[0].selector), "css-duplicate", shape(r[0]))
		}
	}
	// Declarations both sides have, in the order each declares them.
	refOrder, gotOrder := make([]string, 0, len(refDecls)), make([]string, 0, len(gotDecls))
	for _, d := range refDecls {
		if _, ok := gotByKey[key(d)]; ok {
			refOrder = append(refOrder, key(d))
		}
	}
	for _, d := range gotDecls {
		if _, ok := refByKey[key(d)]; ok {
			gotOrder = append(gotOrder, key(d))
		}
	}
	for i := range min(len(refOrder), len(gotOrder)) {
		if refOrder[i] != gotOrder[i] {
			d := refByKey[refOrder[i]][0]
			s.add(cssComponent(d.selector), "css-order", shape(d))
			break
		}
	}
}
