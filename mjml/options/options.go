// Package options contains render options for MJML components
package options

// RenderOpts contains options for MJML rendering
type RenderOpts struct {
	DebugTags   bool // Whether to include debug attributes in output
	InsideGroup bool // Whether the component is being rendered inside a group
}
