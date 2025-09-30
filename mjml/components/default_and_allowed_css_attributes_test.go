package components

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// defaultAttributesData holds the expected default attributes loaded from JSON
type defaultAttributesData map[string]map[string]string

type allowedAttributesData map[string]map[string]string

// loadDefaultAttributesFromJSON loads the expected default attributes from the JSON file
func loadDefaultAttributesFromJSON(t *testing.T) defaultAttributesData {
	t.Helper()

	jsonFile, err := os.ReadFile("testdata/default-css-attrs.json")
	if err != nil {
		t.Fatalf("Failed to read default-css-attrs.json: %v", err)
	}

	var data defaultAttributesData
	if err := json.Unmarshal(jsonFile, &data); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	return data
}

// loadAllowedAttributesFromJSON loads the allowed CSS attributes from the JSON file
func loadAllowedAttributesFromJSON(t *testing.T) allowedAttributesData {
	t.Helper()

	jsonFile, err := os.ReadFile("allowed-css-attributes.json")
	if err != nil {
		t.Fatalf("Failed to read allowed-css-attributes.json: %v", err)
	}

	var data allowedAttributesData
	if err := json.Unmarshal(jsonFile, &data); err != nil {
		t.Fatalf("Failed to parse allowed CSS attributes JSON: %v", err)
	}

	return data
}

// componentConstructors maps component tag names to their constructor functions
var componentConstructors = map[string]func(*parser.MJMLNode, *options.RenderOpts) Component{
	"mj-accordion":       func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJAccordionComponent(n, o) },
	"mj-accordion-text":  func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJAccordionTextComponent(n, o) },
	"mj-accordion-title": func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJAccordionTitleComponent(n, o) },
	"mj-body":            func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJBodyComponent(n, o) },
	"mj-button":          func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJButtonComponent(n, o) },
	"mj-carousel":        func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJCarouselComponent(n, o) },
	"mj-carousel-image":  func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJCarouselImageComponent(n, o) },
	"mj-column":          func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJColumnComponent(n, o) },
	"mj-divider":         func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJDividerComponent(n, o) },
	"mj-group":           func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJGroupComponent(n, o) },
	"mj-hero":            func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJHeroComponent(n, o) },
	"mj-image":           func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJImageComponent(n, o) },
	"mj-navbar":          func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJNavbarComponent(n, o) },
	"mj-navbar-link":     func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJNavbarLinkComponent(n, o) },
	"mj-section":         func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJSectionComponent(n, o) },
	"mj-social":          func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJSocialComponent(n, o) },
	"mj-social-element":  func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJSocialElementComponent(n, o) },
	"mj-spacer":          func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJSpacerComponent(n, o) },
	"mj-table":           func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJTableComponent(n, o) },
	"mj-text":            func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJTextComponent(n, o) },
	"mj-wrapper":         func(n *parser.MJMLNode, o *options.RenderOpts) Component { return NewMJWrapperComponent(n, o) },
}

// createComponent creates a component instance for testing
func createComponent(tagName string) (Component, error) {
	constructor, exists := componentConstructors[tagName]
	if !exists {
		return nil, fmt.Errorf("no constructor found for component: %s", tagName)
	}

	// Create a minimal MJMLNode for testing
	node := &parser.MJMLNode{
		XMLName: xml.Name{Local: tagName},
		Attrs:   []xml.Attr{},
		Text:    "",
	}

	opts := &options.RenderOpts{}

	return constructor(node, opts), nil
}

// TestComponentDefaultAttributes verifies that each MJML component's default attributes
// match the expected values loaded from a predefined JSON file. For each component defined in the
// JSON, it creates an instance, checks that the tag name matches, and compares each
// expected default attribute against the component's actual default attribute value.
// Any missing or mismatched attributes are reported as test failures.
func TestComponentDefaultAttributes(t *testing.T) {
	// Load expected attributes from JSON
	expectedData := loadDefaultAttributesFromJSON(t)

	// Test each component defined in the JSON
	for componentName, expectedAttrs := range expectedData {
		t.Run(componentName, func(t *testing.T) {
			// Create component instance
			component, err := createComponent(componentName)
			if err != nil {
				t.Fatalf("Failed to create component %s: %v", componentName, err)
			}

			// Verify component tag name matches
			if component.GetTagName() != componentName {
				t.Errorf("Component tag name mismatch: expected %s, got %s", componentName, component.GetTagName())
			}

			// Test each expected attribute
			var failures []string
			for attrName, expectedValue := range expectedAttrs {
				actualValue := component.GetDefaultAttribute(attrName)

				if actualValue == "" {
					failures = append(
						failures,
						fmt.Sprintf("  - Missing attribute '%s' (expected: '%s')", attrName, expectedValue),
					)
				} else if actualValue != expectedValue {
					failures = append(failures, fmt.Sprintf("  - Wrong value for '%s': expected '%s', got '%s'", attrName, expectedValue, actualValue))
				}
			}

			// Report failures
			if len(failures) > 0 {
				t.Errorf(
					"Component %s has %d attribute issues:\n%s",
					componentName,
					len(failures),
					strings.Join(failures, "\n"),
				)
			}
		})
	}
}

// TestAllComponentsHaveDefaultAttributes ensures that all components in the codebase
// have corresponding entries in the JSON reference data
func TestAllComponentsHaveDefaultAttributes(t *testing.T) {
	expectedData := loadDefaultAttributesFromJSON(t)

	// Check that all defined constructors have JSON data
	for componentName := range componentConstructors {
		t.Run(componentName, func(t *testing.T) {
			if _, exists := expectedData[componentName]; !exists {
				t.Errorf("Component %s has a constructor but no default attributes defined in JSON", componentName)
			}
		})
	}
}

// TestNoOrphanedJSONEntries ensures that all JSON entries have corresponding components
func TestNoOrphanedJSONEntries(t *testing.T) {
	expectedData := loadDefaultAttributesFromJSON(t)

	// Check that all JSON entries have constructors
	for componentName := range expectedData {
		t.Run(componentName, func(t *testing.T) {
			if _, exists := componentConstructors[componentName]; !exists {
				t.Errorf("JSON defines default attributes for %s but no constructor exists", componentName)
			}
		})
	}
}

func TestComponentAllowedCSSAttributes(t *testing.T) {
	expectedData := loadAllowedAttributesFromJSON(t)

	for componentName, expectedAttrs := range expectedData {
		t.Run(componentName, func(t *testing.T) {
			actualAttrs := AllowedCSSAttributes(componentName)

			if len(actualAttrs) == 0 {
				if len(expectedAttrs) == 0 {
					return
				}
				t.Fatalf("Component %s has no allowed attributes defined", componentName)
			}

			var failures []string

			for attrName, expectedType := range expectedAttrs {
				actualType, exists := actualAttrs[attrName]
				if !exists {
					failures = append(failures, fmt.Sprintf("  - Missing attribute '%s'", attrName))
					continue
				}
				if actualType != expectedType {
					failures = append(failures, fmt.Sprintf("  - Attribute '%s' type mismatch: expected '%s', got '%s'", attrName, expectedType, actualType))
				}
			}

			for attrName := range actualAttrs {
				if _, exists := expectedAttrs[attrName]; !exists {
					failures = append(failures, fmt.Sprintf("  - Unexpected attribute '%s'", attrName))
				}
			}

			if len(failures) > 0 {
				t.Fatalf("Component %s has inconsistent allowed attributes:\n%s", componentName, strings.Join(failures, "\n"))
			}
		})
	}
}
