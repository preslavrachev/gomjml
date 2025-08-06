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
	return CreateComponentWithDepth(node, opts, 0)
}

// CreateComponentWithDepth creates a component from an MJML AST node with depth tracking
func CreateComponentWithDepth(node *parser.MJMLNode, opts *options.RenderOpts, depth int) (Component, error) {
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

	var comp Component
	var err error

	switch tagName {
	case "mjml":
		comp, err = createMJMLComponentWithDepth(node, opts, depth)
		if err != nil {
			return nil, err
		}
	case "mj-head":
		comp = components.NewMJHeadComponent(node, opts)
	case "mj-body":
		comp = components.NewMJBodyComponent(node, opts)
	case "mj-section":
		comp = components.NewMJSectionComponent(node, opts)
	case "mj-column":
		comp = components.NewMJColumnComponent(node, opts)
	case "mj-text":
		comp = components.NewMJTextComponent(node, opts)
	case "mj-button":
		comp = components.NewMJButtonComponent(node, opts)
	case "mj-image":
		comp = components.NewMJImageComponent(node, opts)
	case "mj-title":
		comp = components.NewMJTitleComponent(node, opts)
	case "mj-font":
		comp = components.NewMJFontComponent(node, opts)
	case "mj-wrapper":
		comp = components.NewMJWrapperComponent(node, opts)
	case "mj-divider":
		comp = components.NewMJDividerComponent(node, opts)
	case "mj-social":
		comp = components.NewMJSocialComponent(node, opts)
	case "mj-social-element":
		comp = components.NewMJSocialElementComponent(node, opts)
	case "mj-group":
		comp = components.NewMJGroupComponent(node, opts)
	case "mj-preview":
		comp = components.NewMJPreviewComponent(node, opts)
	case "mj-style":
		comp = components.NewMJStyleComponent(node, opts)
	case "mj-attributes":
		comp = components.NewMJAttributesComponent(node, opts)
	case "mj-all":
		comp = components.NewMJAllComponent(node, opts)
	case "mj-accordion":
		comp = components.NewMJAccordionComponent(node, opts)
	case "mj-accordion-text":
		comp = components.NewMJAccordionTextComponent(node, opts)
	case "mj-accordion-title":
		comp = components.NewMJAccordionTitleComponent(node, opts)
	case "mj-accordion-element":
		comp = components.NewMJAccordionElementComponent(node, opts)
	case "mj-carousel":
		comp = components.NewMJCarouselComponent(node, opts)
	case "mj-carousel-image":
		comp = components.NewMJCarouselImageComponent(node, opts)
	case "mj-hero":
		comp = components.NewMJHeroComponent(node, opts)
	case "mj-navbar":
		comp = components.NewMJNavbarComponent(node, opts)
	case "mj-navbar-link":
		comp = components.NewMJNavbarLinkComponent(node, opts)
	case "mj-spacer":
		comp = components.NewMJSpacerComponent(node, opts)
	case "mj-table":
		comp = components.NewMJTableComponent(node, opts)
	case "mj-raw":
		comp = components.NewMJRawComponent(node, opts)
	default:
		debug.DebugLogError("component", "create-error", "Unknown component type", fmt.Errorf("unknown component: %s", tagName))
		return nil, fmt.Errorf("unknown component: %s", tagName)
	}

	// Set depth on the component
	comp.SetDepth(depth)
	return comp, nil
}

func createMJMLComponent(node *parser.MJMLNode, opts *options.RenderOpts) (*MJMLComponent, error) {
	return createMJMLComponentWithDepth(node, opts, 0)
}

func createMJMLComponentWithDepth(node *parser.MJMLNode, opts *options.RenderOpts, depth int) (*MJMLComponent, error) {
	comp := &MJMLComponent{
		BaseComponent: components.NewBaseComponent(node, opts),
	}

	// Set depth for root MJML component
	comp.SetDepth(depth)

	// Find head and body components
	if headNode := node.FindFirstChild("mj-head"); headNode != nil {
		head := components.NewMJHeadComponent(headNode, opts)
		head.SetDepth(depth + 1)

		// Process head children
		for _, childNode := range headNode.Children {
			if childComponent, err := CreateComponentWithDepth(childNode, opts, depth+2); err == nil {
				head.Children = append(head.Children, childComponent)
			}
		}

		comp.Head = head
	}

	if bodyNode := node.FindFirstChild("mj-body"); bodyNode != nil {
		body := components.NewMJBodyComponent(bodyNode, opts)
		body.SetDepth(depth + 1)
		comp.Body = body

		// Process body children
		for _, childNode := range bodyNode.Children {
			if childComponent, err := CreateComponentWithDepth(childNode, opts, depth+2); err == nil {
				body.Children = append(body.Children, childComponent)
			}
		}

		// Process nested children (sections/wrappers -> columns -> content)
		for _, child := range body.Children {
			switch comp := child.(type) {
			case *components.MJSectionComponent:
				processSectionChildrenWithDepth(comp, opts, depth+3)
			case *components.MJWrapperComponent:
				for _, childNode := range comp.Node.Children {
					if childComponent, err := CreateComponentWithDepth(childNode, opts, depth+3); err == nil {
						comp.Children = append(comp.Children, childComponent)

						// Process wrapper's section children
						if section, ok := childComponent.(*components.MJSectionComponent); ok {
							processSectionChildrenWithDepth(section, opts, depth+4)
						}
					}
				}
			}
		}
	}

	return comp, nil
}

// processComponentChildren recursively processes children of content components
func processComponentChildren(component Component, node *parser.MJMLNode, opts *options.RenderOpts) {
	processComponentChildrenWithDepth(component, node, opts, 5) // Default depth for nested components
}

// processComponentChildrenWithDepth recursively processes children of content components with depth tracking
func processComponentChildrenWithDepth(component Component, node *parser.MJMLNode, opts *options.RenderOpts, childDepth int) {
	// Only process components that can have children
	switch comp := component.(type) {
	case *components.MJSocialComponent:
		// Process social element children
		for _, childNode := range node.Children {
			if childComponent, err := CreateComponentWithDepth(childNode, opts, childDepth); err == nil {
				comp.Children = append(comp.Children, childComponent)
			}
		}
	case *components.MJAccordionComponent:
		// Process accordion element children
		for _, childNode := range node.Children {
			if childComponent, err := CreateComponentWithDepth(childNode, opts, childDepth); err == nil {
				comp.Children = append(comp.Children, childComponent)

				// Process accordion element children (title and text)
				if accordionElement, ok := childComponent.(*components.MJAccordionElementComponent); ok {
					for _, elementChildNode := range childNode.Children {
						if elementChildComponent, err := CreateComponentWithDepth(elementChildNode, opts, childDepth+1); err == nil {
							accordionElement.Children = append(accordionElement.Children, elementChildComponent)
						}
					}
				}
			}
		}
		// Add more component types here as needed
	}
}

// processSectionChildren processes the children of a section component (columns and groups)
func processSectionChildren(section *components.MJSectionComponent, opts *options.RenderOpts) {
	processSectionChildrenWithDepth(section, opts, 3) // Default depth for sections
}

// processSectionChildrenWithDepth processes the children of a section component with depth tracking
func processSectionChildrenWithDepth(section *components.MJSectionComponent, opts *options.RenderOpts, columnDepth int) {
	for _, colNode := range section.Node.Children {
		if colComponent, err := CreateComponentWithDepth(colNode, opts, columnDepth); err == nil {
			section.Children = append(section.Children, colComponent)

			// Handle different column types
			switch col := colComponent.(type) {
			case *components.MJColumnComponent:
				// Process column children
				for _, contentNode := range col.Node.Children {
					if contentComponent, err := CreateComponentWithDepth(contentNode, opts, columnDepth+1); err == nil {
						col.Children = append(col.Children, contentComponent)

						// Process nested children (e.g., social elements within social component)
						processComponentChildrenWithDepth(contentComponent, contentNode, opts, columnDepth+2)
					}
				}
			case *components.MJGroupComponent:
				// Process group children (columns within the group)
				for _, groupChildNode := range col.Node.Children {
					if groupChildComponent, err := CreateComponentWithDepth(groupChildNode, opts, columnDepth+1); err == nil {
						col.Children = append(col.Children, groupChildComponent)

						// Process column children within the group
						if groupColumn, ok := groupChildComponent.(*components.MJColumnComponent); ok {
							for _, contentNode := range groupColumn.Node.Children {
								if contentComponent, err := CreateComponentWithDepth(contentNode, opts, columnDepth+2); err == nil {
									groupColumn.Children = append(groupColumn.Children, contentComponent)

									// Process nested children (e.g., social elements within social component)
									processComponentChildrenWithDepth(contentComponent, contentNode, opts, columnDepth+3)
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
// It creates a strings.Builder writer and passes it to the Component's RenderHTML method.
// This function returns the rendered output as a string, or an error if rendering fails.
// Where possible, it is recommended to use the component's RenderHTML function directly for better performance and flexibility.
func RenderComponentString(c Component) (string, error) {
	var output strings.Builder
	if err := c.RenderHTML(&output); err != nil {
		return "", err
	}
	return output.String(), nil
}

// RenderComponentMJMLString renders the given Component to an MJML string.
// It creates a strings.Builder writer and passes it to the Component's RenderMJML method.
// This function returns the rendered MJML output as a string, or an error if rendering fails.
// Where possible, it is recommended to use the component's RenderMJML function directly for better performance and flexibility.
func RenderComponentMJMLString(c Component) (string, error) {
	var output strings.Builder
	if err := c.RenderMJML(&output); err != nil {
		return "", err
	}
	return output.String(), nil
}
