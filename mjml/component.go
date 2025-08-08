// Package mjml provides the core MJML component system and rendering functionality.
// It converts parsed MJML AST nodes into renderable components that generate HTML output.
package mjml

import (
	"fmt"
	"strings"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/debug"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// Component is an alias for the components.Component interface
type Component = components.Component

// CreateComponent creates a component from an MJML AST node
func CreateComponent(node *parser.MJMLNode, opts *options.RenderOpts) (Component, error) {
	// Ensure opts is not nil and has FontTracker initialized
	if opts == nil {
		opts = &options.RenderOpts{
			FontTracker: options.NewFontTracker(),
		}
	} else if opts.FontTracker == nil {
		opts.FontTracker = options.NewFontTracker()
	}

	tagName := node.GetTagName()

	// Log component creation
	debug.DebugLogWithData("component", "create", "Creating component", map[string]interface{}{
		"tag_name":     tagName,
		"has_children": len(node.Children) > 0,
		"attr_count":   len(node.Attrs),
	})

	switch tagName {
	case "mjml":
		return createMJMLComponent(node, opts)
	case "mj-head":
		return components.NewMJHeadComponent(node, opts), nil
	case "mj-body":
		return components.NewMJBodyComponent(node, opts), nil
	case "mj-section":
		return components.NewMJSectionComponent(node, opts), nil
	case "mj-column":
		return components.NewMJColumnComponent(node, opts), nil
	case "mj-text":
		return components.NewMJTextComponent(node, opts), nil
	case "mj-button":
		return components.NewMJButtonComponent(node, opts), nil
	case "mj-image":
		return components.NewMJImageComponent(node, opts), nil
	case "mj-title":
		return components.NewMJTitleComponent(node, opts), nil
	case "mj-font":
		return components.NewMJFontComponent(node, opts), nil
	case "mj-wrapper":
		return components.NewMJWrapperComponent(node, opts), nil
	case "mj-divider":
		return components.NewMJDividerComponent(node, opts), nil
	case "mj-social":
		return components.NewMJSocialComponent(node, opts), nil
	case "mj-social-element":
		return components.NewMJSocialElementComponent(node, opts), nil
	case "mj-group":
		return components.NewMJGroupComponent(node, opts), nil
	case "mj-preview":
		return components.NewMJPreviewComponent(node, opts), nil
	case "mj-style":
		return components.NewMJStyleComponent(node, opts), nil
	case "mj-attributes":
		return components.NewMJAttributesComponent(node, opts), nil
	case "mj-all":
		return components.NewMJAllComponent(node, opts), nil
	case "mj-accordion":
		return components.NewMJAccordionComponent(node, opts), nil
	case "mj-accordion-text":
		return components.NewMJAccordionTextComponent(node, opts), nil
	case "mj-accordion-title":
		return components.NewMJAccordionTitleComponent(node, opts), nil
	case "mj-accordion-element":
		return components.NewMJAccordionElementComponent(node, opts), nil
	case "mj-carousel":
		return components.NewMJCarouselComponent(node, opts), nil
	case "mj-carousel-image":
		return components.NewMJCarouselImageComponent(node, opts), nil
	case "mj-hero":
		return components.NewMJHeroComponent(node, opts), nil
	case "mj-navbar":
		return components.NewMJNavbarComponent(node, opts), nil
	case "mj-navbar-link":
		return components.NewMJNavbarLinkComponent(node, opts), nil
	case "mj-spacer":
		return components.NewMJSpacerComponent(node, opts), nil
	case "mj-table":
		return components.NewMJTableComponent(node, opts), nil
	case "mj-raw":
		return components.NewMJRawComponent(node, opts), nil
	default:
		debug.DebugLogError("component", "create-error", "Unknown component type", fmt.Errorf("unknown component: %s", tagName))
		return nil, fmt.Errorf("unknown component: %s", tagName)
	}
}

func createMJMLComponent(node *parser.MJMLNode, opts *options.RenderOpts) (*MJMLComponent, error) {
	comp := &MJMLComponent{
		BaseComponent: components.NewBaseComponent(node, opts),
	}

	// Find head and body components
	if headNode := node.FindFirstChild("mj-head"); headNode != nil {
		head := components.NewMJHeadComponent(headNode, opts)

		// Process head children
		for _, childNode := range headNode.Children {
			if childComponent, err := CreateComponent(childNode, opts); err == nil {
				head.Children = append(head.Children, childComponent)
			}
		}

		comp.Head = head
	}

	if bodyNode := node.FindFirstChild("mj-body"); bodyNode != nil {
		body := components.NewMJBodyComponent(bodyNode, opts)
		comp.Body = body

		// Process body children
		for _, childNode := range bodyNode.Children {
			if childComponent, err := CreateComponent(childNode, opts); err == nil {
				body.Children = append(body.Children, childComponent)
			}
		}

		// Process nested children (sections/wrappers -> columns -> content)
		for _, child := range body.Children {
			switch comp := child.(type) {
			case *components.MJSectionComponent:
				processSectionChildren(comp, opts)
			case *components.MJWrapperComponent:
				for _, childNode := range comp.Node.Children {
					if childComponent, err := CreateComponent(childNode, opts); err == nil {
						comp.Children = append(comp.Children, childComponent)

						// Process wrapper's section children
						if section, ok := childComponent.(*components.MJSectionComponent); ok {
							processSectionChildren(section, opts)
						}
					}
				}
			case *components.MJHeroComponent:
				// Process hero children
				processComponentChildren(comp, comp.Node, opts)
			}
		}
	}

	return comp, nil
}

// processComponentChildren recursively processes children of content components
func processComponentChildren(component Component, node *parser.MJMLNode, opts *options.RenderOpts) {
	// Only process components that can have children
	switch comp := component.(type) {
	case *components.MJSocialComponent:
		// Process social element children
		for _, childNode := range node.Children {
			if childComponent, err := CreateComponent(childNode, opts); err == nil {
				comp.Children = append(comp.Children, childComponent)
			}
		}
	case *components.MJAccordionComponent:
		// Process accordion element children
		for _, childNode := range node.Children {
			if childComponent, err := CreateComponent(childNode, opts); err == nil {
				comp.Children = append(comp.Children, childComponent)

				// Process accordion element children (title and text)
				if accordionElement, ok := childComponent.(*components.MJAccordionElementComponent); ok {
					for _, elementChildNode := range childNode.Children {
						if elementChildComponent, err := CreateComponent(elementChildNode, opts); err == nil {
							accordionElement.Children = append(accordionElement.Children, elementChildComponent)
						}
					}
				}
			}
		}
	case *components.MJNavbarComponent:
		// Process navbar link children
		for _, childNode := range node.Children {
			if childComponent, err := CreateComponent(childNode, opts); err == nil {
				comp.Children = append(comp.Children, childComponent)
			}
		}
	case *components.MJHeroComponent:
		// Process hero children
		for _, childNode := range node.Children {
			if childComponent, err := CreateComponent(childNode, opts); err == nil {
				comp.Children = append(comp.Children, childComponent)
			}
		}
		// Add more component types here as needed
	}
}

// processSectionChildren processes the children of a section component (columns and groups)
func processSectionChildren(section *components.MJSectionComponent, opts *options.RenderOpts) {
	for _, colNode := range section.Node.Children {
		if colComponent, err := CreateComponent(colNode, opts); err == nil {
			section.Children = append(section.Children, colComponent)

			// Handle different column types
			switch col := colComponent.(type) {
			case *components.MJColumnComponent:
				// Process column children
				for _, contentNode := range col.Node.Children {
					if contentComponent, err := CreateComponent(contentNode, opts); err == nil {
						col.Children = append(col.Children, contentComponent)

						// Process nested children (e.g., social elements within social component)
						processComponentChildren(contentComponent, contentNode, opts)
					}
				}
			case *components.MJGroupComponent:
				// Process group children (columns within the group)
				for _, groupChildNode := range col.Node.Children {
					if groupChildComponent, err := CreateComponent(groupChildNode, opts); err == nil {
						col.Children = append(col.Children, groupChildComponent)

						// Process column children within the group
						if groupColumn, ok := groupChildComponent.(*components.MJColumnComponent); ok {
							for _, contentNode := range groupColumn.Node.Children {
								if contentComponent, err := CreateComponent(contentNode, opts); err == nil {
									groupColumn.Children = append(groupColumn.Children, contentComponent)

									// Process nested children (e.g., social elements within social component)
									processComponentChildren(contentComponent, contentNode, opts)
								}
							}
						}
					}
				}
			}
		}
	}
}

// RenderComponentString renders the given Component to a string.
// It creates a strings.Builder writer and passes it to the Component's Render method.
// This function returns the rendered output as a string, or an error if rendering fails.
// Where possible, it is recommended to use the component's Render function directly for better performance and flexibility.
func RenderComponentString(c Component) (string, error) {
	var output strings.Builder
	if err := c.Render(&output); err != nil {
		return "", err
	}
	return output.String(), nil
}
