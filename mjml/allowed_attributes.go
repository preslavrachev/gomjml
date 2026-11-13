package mjml

// WithAllowedAttributes accepts attributes that validation would otherwise
// report, such as metadata kept in the MJML source. allow is asked only about
// attributes the tag does not support; returning true accepts one.
func WithAllowedAttributes(allow func(tagName, attrName string) bool) RenderOption {
	return func(opts *RenderOpts) {
		opts.AllowAttribute = allow
	}
}
