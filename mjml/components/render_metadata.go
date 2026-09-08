package components

import "github.com/preslavrachev/gomjml/mjml/constants"

// RenderMetadata holds the document-head-relevant information a full body
// render would otherwise discover as side effects (tracked font families,
// the empty-style-tag requirement). Collecting it in a dedicated pass lets
// the body be rendered exactly once, directly to the caller's writer.
type RenderMetadata struct {
	fontFamilies         map[string]struct{}
	RequireEmptyStyleTag bool
}

// CollectRenderMetadata walks the component tree rooted at root and returns
// the metadata rendering would otherwise accumulate as side effects. It
// calls no Render, GetAttributeWithDefault, getAttribute, or TrackFontFamily
// method, relying instead on each component's pure resolver so the same
// precedence rules rendering uses are never duplicated here. Container
// children are filtered through each component's own renderable-children
// helper (e.g. MJAccordionComponent.renderableElements), so this walker
// never maintains a second, independent interpretation of which children a
// container actually renders.
func CollectRenderMetadata(root Component) RenderMetadata {
	meta := RenderMetadata{}
	collectMetadataFromComponent(root, &meta)
	return meta
}

// FontFamilies returns the resolved font families discovered by the pass, as
// a snapshot slice in unspecified order. Callers needing a stable emission
// order (e.g. font URL generation) must impose their own ordering.
func (m RenderMetadata) FontFamilies() []string {
	families := make([]string, 0, len(m.fontFamilies))
	for family := range m.fontFamilies {
		families = append(families, family)
	}
	return families
}

func (m *RenderMetadata) addFont(family string) {
	if family == "" {
		return
	}
	if m.fontFamilies == nil {
		m.fontFamilies = make(map[string]struct{})
	}
	m.fontFamilies[family] = struct{}{}
}

func collectMetadataChildren(children []Component, meta *RenderMetadata) {
	for _, child := range children {
		collectMetadataFromComponent(child, meta)
	}
}

// collectMetadataFromComponent dispatches on concrete component type. It is
// an exhaustive, manually maintained list: any new component whose Render
// resolves and styles a font-family (directly, like MJTextComponent, or via
// a renderable-children filter, like MJAccordionComponent) must add a case
// here, or its fonts will silently go undetected and never get imported.
func collectMetadataFromComponent(comp Component, meta *RenderMetadata) {
	switch v := comp.(type) {
	case *MJBodyComponent:
		collectMetadataChildren(v.Children, meta)
	case *MJSectionComponent:
		if len(v.Children) == 1 {
			if col, ok := v.Children[0].(*MJColumnComponent); ok && col.requiresSingleColumnSplit() {
				meta.RequireEmptyStyleTag = true
			}
		}
		collectMetadataChildren(v.Children, meta)
	case *MJWrapperComponent:
		collectMetadataChildren(v.Children, meta)
	case *MJGroupComponent:
		collectMetadataChildren(v.renderableChildren(), meta)
	case *MJColumnComponent:
		collectMetadataChildren(v.Children, meta)
	case *MJHeroComponent:
		collectMetadataChildren(v.Children, meta)
	case *MJCarouselComponent:
		collectMetadataChildren(v.Children, meta)

	case *MJTextComponent:
		meta.addFont(v.ResolveAttribute(v, constants.MJMLFontFamily))
	case *MJButtonComponent:
		meta.addFont(v.ResolveAttribute(v, constants.MJMLFontFamily))
	case *MJTableComponent:
		meta.addFont(v.ResolveAttribute(v, constants.MJMLFontFamily))

	case *MJAccordionComponent:
		meta.addFont(v.ResolveAttribute(v, constants.MJMLFontFamily))
		for _, el := range v.renderableElements() {
			collectMetadataFromComponent(el, meta)
		}
	case *MJAccordionElementComponent:
		titleComponent, textComponent := v.renderedTitleAndText()
		if titleComponent != nil {
			meta.addFont(titleComponent.explicitFontFamily())
		}
		if textComponent != nil {
			meta.addFont(textComponent.explicitFontFamily())
		}

	case *MJNavbarComponent:
		if hamburger := v.ResolveAttribute(v, "hamburger"); hamburger != "" {
			meta.addFont(v.ResolveAttribute(v, "ico-font-family"))
		}
		for _, link := range v.renderableLinks() {
			collectMetadataFromComponent(link, meta)
		}
	case *MJNavbarLinkComponent:
		meta.addFont(v.ResolveAttribute(v, constants.MJMLFontFamily))

	case *MJSocialComponent:
		if v.hasTextContent() {
			meta.addFont(v.ResolveAttribute(v, constants.MJMLFontFamily))
		}
		mode := v.ResolveAttribute(v, constants.MJMLMode)
		for _, child := range v.Children {
			if elem, ok := child.(*MJSocialElementComponent); ok {
				elem.InheritFromParent(v)
				collectSocialElementFont(elem, mode, meta)
			}
		}
	}
}

// collectSocialElementFont reproduces the reachability rule
// MJSocialElementComponent.Render uses to decide whether it ever styles a
// font-family: the element must resolve a non-empty src (otherwise Render
// returns before touching font-family) and carry text content in the mode
// it will render in (vertical uses Node.Text, horizontal uses the mixed
// content).
func collectSocialElementFont(c *MJSocialElementComponent, mode string, meta *RenderMetadata) {
	if c.resolveAttribute("src") == "" {
		return
	}

	var textContent string
	if mode == "vertical" {
		textContent = c.Node.Text
	} else {
		textContent = c.Node.GetMixedContent()
	}
	if textContent == "" {
		return
	}

	meta.addFont(c.resolveAttribute(constants.MJMLFontFamily))
}
