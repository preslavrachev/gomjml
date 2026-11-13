package mjml

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

// prettyHTML prints a document one tag, text run or CSS rule per line, indented by depth, so
// that a line diff of two outputs shows markup differences rather than formatting. Both sides
// go through normalizeForComparison first, which masks generated ids. Attributes and inline style
// declarations are sorted, as the comparison ignores their order; the semantic differences
// report the orders that matter. The inside of each conditional comment is printed the same way, with
// one depth shared across comments, since MSO tables open in one comment and close in another.
func prettyHTML(document string) []string {
	p := &prettyPrinter{}
	p.markup(normalizeForComparison(document))
	return p.lines
}

type prettyPrinter struct {
	lines []string
	depth int
	style bool
	pre   int // open <pre> elements, whose text is printed quoted, whitespace and all
}

var conditionalComment = regexp.MustCompile(`(?s)^(\[if [^\]]*\]>)(.*?)(<!\[endif\])?$`)

func (p *prettyPrinter) line(text string) {
	p.lines = append(p.lines, strings.Repeat("  ", max(p.depth, 0))+text)
}

func (p *prettyPrinter) markup(markup string) {
	tokenizer := html.NewTokenizer(strings.NewReader(markup))
	for {
		switch tokenizer.Next() {
		case html.ErrorToken:
			return
		case html.DoctypeToken:
			p.line("<!DOCTYPE " + string(tokenizer.Text()) + ">")
		case html.StartTagToken:
			token := tokenizer.Token()
			p.line(prettyTag(token))
			if !voidElements[token.Data] {
				p.depth++
			}
			p.style = token.Data == "style"
			if token.Data == "pre" {
				p.pre++
			}
		case html.SelfClosingTagToken:
			p.line(prettyTag(tokenizer.Token()))
		case html.EndTagToken:
			token := tokenizer.Token()
			if !voidElements[token.Data] {
				p.depth--
			}
			p.style = false
			if token.Data == "pre" && p.pre > 0 {
				p.pre--
			}
			p.line("</" + token.Data + ">")
		case html.TextToken:
			text := string(tokenizer.Text())
			if p.style {
				for _, rule := range cssLines(text) {
					p.line(rule)
				}
			} else if p.pre > 0 {
				p.line(strconv.Quote(html.UnescapeString(text)))
			} else if collapsed := strings.Join(strings.Fields(html.UnescapeString(text)), " "); collapsed != "" {
				p.line(collapsed)
			}
		case html.CommentToken:
			data := string(tokenizer.Text())
			if match := conditionalComment.FindStringSubmatch(data); match != nil {
				p.line("<!--" + match[1])
				p.markup(match[2])
				if match[3] != "" {
					p.line("<![endif]-->")
				}
			} else {
				p.line("<!--" + strings.Join(strings.Fields(data), " ") + "-->")
			}
		}
	}
}

var voidElements = map[string]bool{
	"area": true, "base": true, "br": true, "col": true, "embed": true, "hr": true, "img": true,
	"input": true, "link": true, "meta": true, "source": true, "track": true, "wbr": true,
}

func prettyTag(token html.Token) string {
	var attributes []string
	for _, attr := range token.Attr {
		value := attr.Val
		if attr.Key == "style" {
			var declarations []string
			for declaration := range strings.SplitSeq(value, ";") {
				if property, v, found := strings.Cut(declaration, ":"); found {
					declarations = append(declarations, strings.TrimSpace(property)+": "+strings.TrimSpace(v))
				}
			}
			slices.Sort(declarations)
			value = strings.Join(declarations, "; ")
		}
		attributes = append(attributes, fmt.Sprintf("%s=%q", attr.Key, value))
	}
	slices.Sort(attributes)
	if len(attributes) == 0 {
		return "<" + token.Data + ">"
	}
	return "<" + token.Data + " " + strings.Join(attributes, " ") + ">"
}

// cssSpace finds whitespace that cannot change a style sheet, which the comparison ignores too.
var cssSpace = regexp.MustCompile(`\s*([{};:,>+~!()])\s*`)

// cssLines splits a style sheet into one line per rule or at-rule boundary, without the
// whitespace that cannot matter.
func cssLines(css string) []string {
	var lines []string
	var current strings.Builder
	flush := func() {
		if text := cssSpace.ReplaceAllString(strings.Join(strings.Fields(current.String()), " "), "$1"); text != "" {
			lines = append(lines, text)
		}
		current.Reset()
	}
	for _, r := range css {
		current.WriteRune(r)
		if r == '{' && strings.HasPrefix(strings.TrimSpace(current.String()), "@media") || r == '}' {
			flush()
		}
	}
	flush()
	return lines
}

// unifiedDiff returns the hunks of a unified diff between two line slices, with context lines.
func unifiedDiff(from, to []string, context int) []diffHunk {
	ops := diffLines(from, to)
	var hunks []diffHunk
	for i := 0; i < len(ops); {
		if ops[i].kind == ' ' {
			i++
			continue
		}
		start := max(i-context, 0)
		for start < i && ops[start].kind != ' ' {
			start++
		}
		end := i
		for end < len(ops) {
			if ops[end].kind != ' ' {
				end++
				continue
			}
			run := end
			for run < len(ops) && ops[run].kind == ' ' {
				run++
			}
			if run == len(ops) || run-end > 2*context {
				end = min(end+context, len(ops))
				break
			}
			end = run
		}
		hunk := diffHunk{fromLine: ops[start].from + 1, toLine: ops[start].to + 1}
		for _, op := range ops[start:end] {
			hunk.lines = append(hunk.lines, string(op.kind)+op.text)
		}
		hunks = append(hunks, hunk)
		i = end
	}
	return hunks
}

type diffHunk struct {
	fromLine, toLine int
	lines            []string
}

func (h diffHunk) String() string {
	var from, to int
	for _, line := range h.lines {
		if line[0] != '+' {
			from++
		}
		if line[0] != '-' {
			to++
		}
	}
	return fmt.Sprintf("@@ -%d,%d +%d,%d @@\n%s", h.fromLine, from, h.toLine, to, strings.Join(h.lines, "\n"))
}

func formatUnifiedDiff(fromName, toName string, hunks []diffHunk) string {
	if len(hunks) == 0 {
		return ""
	}
	var b strings.Builder
	fmt.Fprintf(&b, "--- %s\n+++ %s\n", fromName, toName)
	for _, hunk := range hunks {
		b.WriteString(hunk.String())
		b.WriteString("\n")
	}
	return b.String()
}

type diffOp struct {
	kind     byte // ' ', '-' or '+'
	text     string
	from, to int // line indexes before this op, 0-based
}

// diffLines computes a shortest edit script with Myers' O(ND) algorithm. It keeps, for each
// edit distance d, only the diagonals -d-1..d+1 that step d reads, so memory grows with D².
func diffLines(a, b []string) []diffOp {
	n, m := len(a), len(b)
	maxD := n + m
	offset := maxD + 1
	v := make([]int, 2*maxD+3)
	var trace [][]int
	for d := 0; d <= maxD; d++ {
		trace = append(trace, slices.Clone(v[offset-d-1:offset+d+2]))
		for k := -d; k <= d; k += 2 {
			var x int
			if k == -d || (k != d && v[offset+k-1] < v[offset+k+1]) {
				x = v[offset+k+1]
			} else {
				x = v[offset+k-1] + 1
			}
			y := x - k
			for x < n && y < m && a[x] == b[y] {
				x++
				y++
			}
			v[offset+k] = x
			if x >= n && y >= m {
				return backtrack(trace, a, b, d)
			}
		}
	}
	return nil
}

func backtrack(trace [][]int, a, b []string, d int) []diffOp {
	var ops []diffOp
	x, y := len(a), len(b)
	for ; d >= 0; d-- {
		at := func(k int) int { return trace[d][k+d+1] }
		k := x - y
		var prevK int
		if k == -d || (k != d && at(k-1) < at(k+1)) {
			prevK = k + 1
		} else {
			prevK = k - 1
		}
		prevX := at(prevK)
		prevY := prevX - prevK
		if d == 0 {
			prevX, prevY = 0, 0
		}
		for x > prevX && y > prevY {
			x--
			y--
			ops = append(ops, diffOp{kind: ' ', text: a[x], from: x, to: y})
		}
		if d > 0 {
			if x == prevX {
				y--
				ops = append(ops, diffOp{kind: '+', text: b[y], from: x, to: y})
			} else {
				x--
				ops = append(ops, diffOp{kind: '-', text: a[x], from: x, to: y})
			}
		}
	}
	slices.Reverse(ops)
	return ops
}
