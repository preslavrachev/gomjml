package mjml

import (
	"strings"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/options"
)

type inlineCSSRule struct {
	selectors    []string
	declarations []options.InlineStyle
}

// collectInlineClassStyles parses mj-style components with inline="inline" and
// returns a map of css-class names to their ordered CSS declarations.
func collectInlineClassStyles(head *components.MJHeadComponent, opts *options.RenderOpts) map[string][]options.InlineStyle {
	if opts != nil {
		opts.SkipInlineStylesInHead = false
	}
	if head == nil {
		return nil
	}

	classStyles := make(map[string][]options.InlineStyle)
	inlineStyleCount := 0
	inlineStyleHasNewline := false
	for _, child := range head.Children {
		styleComp, ok := child.(*components.MJStyleComponent)
		if !ok {
			continue
		}
		inlineAttr := ""
		if attr := styleComp.GetAttribute("inline"); attr != nil {
			inlineAttr = strings.ToLower(strings.TrimSpace(*attr))
		}
		if inlineAttr != "inline" {
			continue
		}

		inlineStyleCount++
		trimmed := strings.TrimSpace(styleComp.Node.Text)
		if strings.Contains(trimmed, "\n") {
			inlineStyleHasNewline = true
		}

		rules := parseInlineCSSRules(styleComp.Node.Text)
		for _, rule := range rules {
			for _, selector := range rule.selectors {
				if className, ok := extractInlineClass(selector); ok {
					classStyles[className] = append(classStyles[className], rule.declarations...)
				}
			}
		}
	}

	if opts != nil {
		opts.SkipInlineStylesInHead = inlineStyleCount == 1 && inlineStyleHasNewline
	}

	if len(classStyles) == 0 {
		return nil
	}
	return classStyles
}

func parseInlineCSSRules(cssText string) []inlineCSSRule {
	text := strings.TrimSpace(cssText)
	if text == "" {
		return nil
	}

	var rules []inlineCSSRule
	for len(text) > 0 {
		start := strings.Index(text, "{")
		if start == -1 {
			break
		}
		selectorPart := strings.TrimSpace(text[:start])
		text = text[start+1:]

		end := strings.Index(text, "}")
		var declarationsPart string
		if end == -1 {
			declarationsPart = text
			text = ""
		} else {
			declarationsPart = text[:end]
			text = text[end+1:]
		}

		selectors := parseInlineSelectors(selectorPart)
		declarations := parseInlineDeclarations(declarationsPart)
		if len(selectors) == 0 || len(declarations) == 0 {
			continue
		}
		rules = append(rules, inlineCSSRule{selectors: selectors, declarations: declarations})
	}
	return rules
}

func parseInlineSelectors(selectorPart string) []string {
	if selectorPart == "" {
		return nil
	}
	parts := strings.Split(selectorPart, ",")
	selectors := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			selectors = append(selectors, trimmed)
		}
	}
	return selectors
}

func parseInlineDeclarations(declarationsPart string) []options.InlineStyle {
	parts := strings.Split(declarationsPart, ";")
	declarations := make([]options.InlineStyle, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		colon := strings.Index(trimmed, ":")
		if colon == -1 {
			continue
		}
		property := strings.TrimSpace(trimmed[:colon])
		value := strings.TrimSpace(trimmed[colon+1:])
		if property == "" || value == "" {
			continue
		}
		declarations = append(declarations, options.InlineStyle{Property: property, Value: value})
	}
	return declarations
}

func extractInlineClass(selector string) (string, bool) {
	trimmed := strings.TrimSpace(selector)
	if !strings.HasPrefix(trimmed, ".") {
		return "", false
	}

	withoutDot := trimmed[1:]
	end := len(withoutDot)
	for i, r := range withoutDot {
		switch r {
		case ' ', '\t', '\n', '\r', '.', '#', ':', '>', '+', '~', '[':
			end = i
			goto done
		}
	}

done:
	className := strings.TrimSpace(withoutDot[:end])
	if className == "" {
		return "", false
	}
	return className, true
}
