// Package mjml provides the core MJML component system and rendering functionality.
// It converts parsed MJML AST nodes into renderable components that generate HTML output.
package mjml

import (
	"fmt"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/parser"
)

// Component is an alias for the components.Component interface
type Component = components.Component

// CreateComponent creates a component from an MJML AST node
func CreateComponent(node *parser.MJMLNode) (Component, error) {
	tagName := node.GetTagName()

	switch tagName {
	case "mjml":
		return createMJMLComponent(node)
	case "mj-head":
		return components.NewMJHeadComponent(node), nil
	case "mj-body":
		return components.NewMJBodyComponent(node), nil
	case "mj-section":
		return components.NewMJSectionComponent(node), nil
	case "mj-column":
		return components.NewMJColumnComponent(node), nil
	case "mj-text":
		return components.NewMJTextComponent(node), nil
	case "mj-button":
		return components.NewMJButtonComponent(node), nil
	case "mj-image":
		return components.NewMJImageComponent(node), nil
	case "mj-title":
		return components.NewMJTitleComponent(node), nil
	case "mj-font":
		return components.NewMJFontComponent(node), nil
	case "mj-wrapper":
		return components.NewMJWrapperComponent(node), nil
	case "mj-divider":
		return components.NewMJDividerComponent(node), nil
	case "mj-social":
		return components.NewMJSocialComponent(node), nil
	case "mj-social-element":
		return components.NewMJSocialElementComponent(node), nil
	case "mj-group":
		return components.NewMJGroupComponent(node), nil
	case "mj-preview":
		return components.NewMJPreviewComponent(node), nil
	case "mj-style":
		return components.NewMJStyleComponent(node), nil
	case "mj-attributes":
		return components.NewMJAttributesComponent(node), nil
	case "mj-all":
		return components.NewMJAllComponent(node), nil
	default:
		return nil, fmt.Errorf("unknown component: %s", tagName)
	}
}

func createMJMLComponent(node *parser.MJMLNode) (*MJMLComponent, error) {
	comp := &MJMLComponent{
		BaseComponent: components.NewBaseComponent(node),
	}

	// Find head and body components
	if headNode := node.FindFirstChild("mj-head"); headNode != nil {
		head := components.NewMJHeadComponent(headNode)

		// Process head children
		for _, childNode := range headNode.Children {
			if childComponent, err := CreateComponent(childNode); err == nil {
				head.Children = append(head.Children, childComponent)
			}
		}

		comp.Head = head
	}

	if bodyNode := node.FindFirstChild("mj-body"); bodyNode != nil {
		body := components.NewMJBodyComponent(bodyNode)
		comp.Body = body

		// Process body children
		for _, childNode := range bodyNode.Children {
			if childComponent, err := CreateComponent(childNode); err == nil {
				body.Children = append(body.Children, childComponent)
			}
		}

		// Process nested children (sections/wrappers -> columns -> content)
		for _, child := range body.Children {
			switch comp := child.(type) {
			case *components.MJSectionComponent:
				for _, colNode := range comp.Node.Children {
					if colComponent, err := CreateComponent(colNode); err == nil {
						comp.Children = append(comp.Children, colComponent)

						// Process column children
						if column, ok := colComponent.(*components.MJColumnComponent); ok {
							for _, contentNode := range column.Node.Children {
								if contentComponent, err := CreateComponent(contentNode); err == nil {
									column.Children = append(column.Children, contentComponent)
								}
							}
						}
					}
				}
			case *components.MJWrapperComponent:
				for _, childNode := range comp.Node.Children {
					if childComponent, err := CreateComponent(childNode); err == nil {
						comp.Children = append(comp.Children, childComponent)

						// Process wrapper's section children
						if section, ok := childComponent.(*components.MJSectionComponent); ok {
							for _, colNode := range section.Node.Children {
								if colComponent, err := CreateComponent(colNode); err == nil {
									section.Children = append(section.Children, colComponent)

									// Process column children
									if column, ok := colComponent.(*components.MJColumnComponent); ok {
										for _, contentNode := range column.Node.Children {
											if contentComponent, err := CreateComponent(contentNode); err == nil {
												column.Children = append(column.Children, contentComponent)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return comp, nil
}
