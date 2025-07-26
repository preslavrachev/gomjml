// Package mjml provides the core MJML component system and rendering functionality.
// It converts parsed MJML AST nodes into renderable components that generate HTML output.
package mjml

import (
	"fmt"

	"github.com/preslavrachev/gomjml/mjml/components"
	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// Component is an alias for the components.Component interface
type Component = components.Component

// CreateComponent creates a component from an MJML AST node
func CreateComponent(node *parser.MJMLNode, opts *options.RenderOpts) (Component, error) {
	tagName := node.GetTagName()

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
	default:
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
				for _, colNode := range comp.Node.Children {
					if colComponent, err := CreateComponent(colNode, opts); err == nil {
						comp.Children = append(comp.Children, colComponent)

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
			case *components.MJWrapperComponent:
				for _, childNode := range comp.Node.Children {
					if childComponent, err := CreateComponent(childNode, opts); err == nil {
						comp.Children = append(comp.Children, childComponent)

						// Process wrapper's section children
						if section, ok := childComponent.(*components.MJSectionComponent); ok {
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
					}
				}
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
		// Add more component types here as needed
	}
}
