package globals

import (
	"github.com/preslavrachev/gomjml/parser"
)

// GlobalAttributes stores global attribute definitions from mj-attributes
type GlobalAttributes struct {
	// all stores mj-all global attributes that apply to all components
	all map[string]string
	// componentDefaults stores component-specific defaults (e.g., mj-text attributes)
	componentDefaults map[string]map[string]string
}

// NewGlobalAttributes creates a new global attributes store
func NewGlobalAttributes() *GlobalAttributes {
	return &GlobalAttributes{
		all:               make(map[string]string),
		componentDefaults: make(map[string]map[string]string),
	}
}

// ProcessAttributesFromHead processes mj-attributes from the head component
func (ga *GlobalAttributes) ProcessAttributesFromHead(headNode *parser.MJMLNode) {
	if headNode == nil {
		return
	}

	// Find mj-attributes elements
	for _, child := range headNode.Children {
		if child.XMLName.Local == "mj-attributes" {
			ga.processAttributesElement(child)
		}
	}
}

// processAttributesElement processes a single mj-attributes element
func (ga *GlobalAttributes) processAttributesElement(attributesNode *parser.MJMLNode) {
	for _, child := range attributesNode.Children {
		tagName := child.XMLName.Local

		if tagName == "mj-all" {
			// Process mj-all - applies to all components
			for _, attr := range child.Attrs {
				ga.all[attr.Name.Local] = attr.Value
			}
		} else {
			// Process component-specific defaults (e.g., mj-text)
			if ga.componentDefaults[tagName] == nil {
				ga.componentDefaults[tagName] = make(map[string]string)
			}
			for _, attr := range child.Attrs {
				ga.componentDefaults[tagName][attr.Name.Local] = attr.Value
			}
		}
	}
}

// GetGlobalAttribute gets a global attribute value for a component
func (ga *GlobalAttributes) GetGlobalAttribute(componentName, attrName string) string {
	// Check component-specific defaults first
	if componentDefaults, exists := ga.componentDefaults[componentName]; exists {
		if value, exists := componentDefaults[attrName]; exists {
			return value
		}
	}

	// Check mj-all defaults
	if value, exists := ga.all[attrName]; exists {
		return value
	}

	return ""
}

// Global instance (will be set during rendering)
var instance *GlobalAttributes

// SetGlobalAttributes sets the global attributes instance
func SetGlobalAttributes(ga *GlobalAttributes) {
	instance = ga
}

// GetGlobalAttribute is a package-level function to access global attributes
func GetGlobalAttribute(componentName, attrName string) string {
	if instance == nil {
		return ""
	}
	return instance.GetGlobalAttribute(componentName, attrName)
}
