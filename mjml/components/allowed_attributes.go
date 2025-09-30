package components

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	_ "embed"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

//go:embed allowed-css-attributes.json
var allowedCSSAttributesJSON []byte

var (
	allowedAttributesOnce sync.Once
	allowedAttributes     map[string]map[string]string
	allowedAttributeSets  map[string]map[string]struct{}
	allowedAttributesErr  error
)

func ensureAllowedAttributesLoaded() {
	allowedAttributesOnce.Do(func() {
		allowedAttributes = make(map[string]map[string]string)
		allowedAttributeSets = make(map[string]map[string]struct{})

		if err := json.Unmarshal(allowedCSSAttributesJSON, &allowedAttributes); err != nil {
			allowedAttributesErr = fmt.Errorf("failed to parse allowed CSS attributes: %w", err)
			return
		}

		for component, attrs := range allowedAttributes {
			set := make(map[string]struct{}, len(attrs))
			for attr := range attrs {
				set[attr] = struct{}{}
			}
			allowedAttributeSets[component] = set
		}
	})

	if allowedAttributesErr != nil {
		panic(allowedAttributesErr)
	}
}

// AllowedCSSAttributes returns a copy of the allowed attribute map for a component.
func AllowedCSSAttributes(tagName string) map[string]string {
	ensureAllowedAttributesLoaded()
	attrs, ok := allowedAttributes[tagName]
	if !ok {
		return nil
	}

	copyAttrs := make(map[string]string, len(attrs))
	for k, v := range attrs {
		copyAttrs[k] = v
	}
	return copyAttrs
}

func getAllowedAttributeSet(tagName string) (map[string]struct{}, bool) {
	ensureAllowedAttributesLoaded()
	set, ok := allowedAttributeSets[tagName]
	return set, ok
}

var globalAllowedAttributes = map[string]struct{}{
	"mj-class":  {},
	"css-class": {},
	"class":     {},
}

func isGloballyAllowedAttribute(attrName string) bool {
	if attrName == "" {
		return false
	}
	if strings.HasPrefix(attrName, "data-") || strings.HasPrefix(attrName, "aria-") {
		return true
	}
	_, ok := globalAllowedAttributes[attrName]
	return ok
}

func validateComponentAttributes(node *parser.MJMLNode, opts *options.RenderOpts) {
	if node == nil || opts == nil || opts.InvalidAttributeReporter == nil {
		return
	}

	tagName := node.GetTagName()
	allowedSet, ok := getAllowedAttributeSet(tagName)
	if !ok {
		return
	}

	line := node.GetLineNumber()
	for _, attr := range node.Attrs {
		name := attr.Name.Local
		if isGloballyAllowedAttribute(name) {
			continue
		}
		if _, exists := allowedSet[name]; exists {
			continue
		}
		opts.InvalidAttributeReporter(tagName, name, line)
	}
}
