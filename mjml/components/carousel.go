package components

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// Global pseudo-random state for unique carousel IDs. We use a deterministic
// linear congruential generator to mirror MJML's random hex identifiers while
// keeping test runs repeatable.
var carouselIDState uint64

const (
	// Default seed matches the initial value used by MJML's JavaScript
	// compiler (see packages/mjml-carousel/src/index.js). Starting from the
	// same point keeps our deterministic LCG aligned with the canonical HTML
	// output produced by the reference implementation.
	carouselIDSeed = 0xf01ab44896143632
	// Parameters from PCG-XSH-RR. Any full-period LCG works; these provide
	// good distribution while remaining lightweight.
	carouselIDMultiplier = 6364136223846793005
	carouselIDIncrement  = 1442695040888963407
)

// ResetCarouselIDCounter resets the global counter for deterministic testing
func ResetCarouselIDCounter() {
	atomic.StoreUint64(&carouselIDState, 0)
}

// MJCarouselComponent represents the mj-carousel component
type MJCarouselComponent struct {
	*BaseComponent
	Children []Component
	id       string
}

func NewMJCarouselComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJCarouselComponent {
	return &MJCarouselComponent{
		BaseComponent: NewBaseComponent(node, opts),
		Children:      []Component{}, // Will be populated by the component factory
	}
}

func (c *MJCarouselComponent) Render(w io.StringWriter) error {
	// Generate unique carousel ID
	carouselID := c.generateCarouselID()

	// Get carousel images from children
	carouselImages := c.getCarouselImages()
	if len(carouselImages) == 0 {
		return fmt.Errorf("mj-carousel requires at least one mj-carousel-image child")
	}

	// Create table row wrapper (required for content components)
	if _, err := w.WriteString("<tr>"); err != nil {
		return err
	}

	// Create table cell with proper alignment and CSS classes
	align := c.GetAttributeWithDefault(c, "align")
	classAttr := c.BuildClassAttribute()

	if _, err := w.WriteString(`<td align="`); err != nil {
		return err
	}
	if _, err := w.WriteString(align); err != nil {
		return err
	}
	if _, err := w.WriteString(`"`); err != nil {
		return err
	}
	if classAttr != "" {
		if _, err := w.WriteString(` class="`); err != nil {
			return err
		}
		if _, err := w.WriteString(classAttr); err != nil {
			return err
		}
		if _, err := w.WriteString(`"`); err != nil {
			return err
		}
	}
	inlineStyle := ""
	if classAttr != "" {
		inlineStyle = c.BuildInlineStyleString(classAttr)
	}
	if _, err := w.WriteString(` style="`); err != nil {
		return err
	}
	if _, err := w.WriteString(inlineStyle); err != nil {
		return err
	}
	if _, err := w.WriteString("font-size:0px;word-break:break-word;\">"); err != nil {
		return err
	}

	// Render main carousel content
	if err := c.renderCarouselContent(w, carouselID, carouselImages); err != nil {
		return err
	}

	// Render MSO fallback
	if err := c.renderMSOFallback(w, carouselImages); err != nil {
		return err
	}

	// Close table cell and row
	if _, err := w.WriteString("</td></tr>"); err != nil {
		return err
	}

	return nil
}

func (c *MJCarouselComponent) GetTagName() string {
	return "mj-carousel"
}

// GenerateCSS generates the CSS for this carousel component
func (c *MJCarouselComponent) GenerateCSS() string {
	// Generate unique carousel ID
	carouselID := c.generateCarouselID()

	// Get carousel images from children
	carouselImages := c.getCarouselImages()
	if len(carouselImages) == 0 {
		return ""
	}

	return c.buildCarouselCSS(carouselID, len(carouselImages))
}

func (c *MJCarouselComponent) GetDefaultAttribute(name string) string {
	// TODO: Consider more performant approaches to attribute matching than switch statements,
	// such as static map[string]string lookups or compile-time generated code for components
	// with many default attributes (10+ attributes). Switch statements may have O(n) lookup
	// time while map lookups are O(1) average case.
	switch name {
	case "align":
		return "center"
	case "border-radius":
		return "6px"
	case "icon-width":
		return "44px"
	case "left-icon":
		return "https://i.imgur.com/xTh3hln.png"
	case "right-icon":
		return "https://i.imgur.com/os7o9kz.png"
	case "tb-border":
		return "2px solid transparent"
	case "tb-border-radius":
		return "6px"
	case "tb-hover-border-color":
		return "#fead0d"
	case "tb-selected-border-color":
		return "#cccccc"
	case "tb-width":
		return "110px"
	case "thumbnails":
		return "visible"
	default:
		return ""
	}
}

// MJCarouselImageComponent represents the mj-carousel-image component
type MJCarouselImageComponent struct {
	*BaseComponent
}

func NewMJCarouselImageComponent(node *parser.MJMLNode, opts *options.RenderOpts) *MJCarouselImageComponent {
	return &MJCarouselImageComponent{
		BaseComponent: NewBaseComponent(node, opts),
	}
}

func (c *MJCarouselImageComponent) Render(w io.StringWriter) error {
	// mj-carousel-image doesn't render standalone - it's rendered by parent mj-carousel
	return fmt.Errorf("mj-carousel-image cannot render standalone - must be child of mj-carousel")
}

func (c *MJCarouselImageComponent) GetTagName() string {
	return "mj-carousel-image"
}

func (c *MJCarouselImageComponent) GetDefaultAttribute(name string) string {
	switch name {
	case "target":
		return "_blank"
	default:
		return ""
	}
}

// buildCarouselCSS builds the carousel CSS content as a string
func (c *MJCarouselComponent) buildCarouselCSS(carouselID string, imageCount int) string {
	iconWidth := c.GetAttributeWithDefault(c, "icon-width")
	tbHoverBorderColor := c.GetAttributeWithDefault(c, "tb-hover-border-color")
	tbSelectedBorderColor := c.GetAttributeWithDefault(c, "tb-selected-border-color")

	var css strings.Builder

	repeat := func(count int) string {
		if count <= 0 {
			return ""
		}
		return strings.Repeat("+ * ", count)
	}

	writeSelectorBlock := func(selectors []string, joiner, body string) {
		if len(selectors) == 0 {
			return
		}

		css.WriteString("    ")
		css.WriteString(strings.Join(selectors, joiner))
		css.WriteString(" {\n")
		css.WriteString(body)
		css.WriteString("\n    }\n\n")
	}

	css.WriteString(".mj-carousel {\n")
	css.WriteString("      -webkit-user-select: none;\n")
	css.WriteString("      -moz-user-select: none;\n")
	css.WriteString("      user-select: none;\n")
	css.WriteString("    }\n\n")

	writeSelectorBlock(
		[]string{fmt.Sprintf(".mj-carousel-%s-icons-cell", carouselID)},
		",\n    ",
		fmt.Sprintf("      display: table-cell !important;\n      width: %s !important;", iconWidth),
	)

	css.WriteString("    .mj-carousel-radio,\n")
	css.WriteString("    .mj-carousel-next,\n")
	css.WriteString("    .mj-carousel-previous {\n")
	css.WriteString("      display: none !important;\n")
	css.WriteString("    }\n\n")

	css.WriteString("    .mj-carousel-thumbnail,\n")
	css.WriteString("    .mj-carousel-next,\n")
	css.WriteString("    .mj-carousel-previous {\n")
	css.WriteString("      touch-action: manipulation;\n")
	css.WriteString("    }\n\n")

	hideSelectors := make([]string, imageCount)
	for i := 0; i < imageCount; i++ {
		hideSelectors[i] = fmt.Sprintf(
			".mj-carousel-%s-radio:checked %s+ .mj-carousel-content .mj-carousel-image",
			carouselID,
			repeat(i),
		)
	}
	writeSelectorBlock(hideSelectors, ",", "      display: none !important;")

	showSelectors := make([]string, imageCount)
	for i := 0; i < imageCount; i++ {
		showSelectors[i] = fmt.Sprintf(
			".mj-carousel-%s-radio-%d:checked %s+ .mj-carousel-content .mj-carousel-image-%d",
			carouselID,
			i+1,
			repeat(imageCount-i-1),
			i+1,
		)
	}
	writeSelectorBlock(showSelectors, ",", "      display: block !important;")

	nextSelectors := make([]string, imageCount)
	previousSelectors := make([]string, imageCount)
	for i := 0; i < imageCount; i++ {
		padding := repeat(imageCount - i - 1)

		nextSelectors[i] = fmt.Sprintf(
			".mj-carousel-%s-radio-%d:checked %s+ .mj-carousel-content .mj-carousel-next-%d",
			carouselID,
			i+1,
			padding,
			(i+1)%imageCount+1,
		)
		previousSelectors[i] = fmt.Sprintf(
			".mj-carousel-%s-radio-%d:checked %s+ .mj-carousel-content .mj-carousel-previous-%d",
			carouselID,
			i+1,
			padding,
			(i-1+imageCount)%imageCount+1,
		)
	}

	css.WriteString("    .mj-carousel-previous-icons,\n")
	css.WriteString("    .mj-carousel-next-icons,\n")
	css.WriteString("    ")
	css.WriteString(strings.Join(nextSelectors, ","))
	css.WriteString(",\n")
	css.WriteString("    ")
	css.WriteString(strings.Join(previousSelectors, ","))
	css.WriteString(" {\n")
	css.WriteString("      display: block !important;\n")
	css.WriteString("    }\n\n")

	thumbSelectors := make([]string, imageCount)
	for i := 0; i < imageCount; i++ {
		thumbSelectors[i] = fmt.Sprintf(
			".mj-carousel-%s-radio-%d:checked %s+ .mj-carousel-content .mj-carousel-%s-thumbnail-%d",
			carouselID,
			i+1,
			repeat(imageCount-i-1),
			carouselID,
			i+1,
		)
	}
	writeSelectorBlock(
		thumbSelectors,
		",",
		fmt.Sprintf("      border-color: %s !important;", tbSelectedBorderColor),
	)

	writeSelectorBlock(
		[]string{".mj-carousel-image img + div", ".mj-carousel-thumbnail img + div"},
		",\n    ",
		"      display: none !important;",
	)

	hoverSelectors := make([]string, 0, imageCount)
	for i := imageCount - 1; i >= 0; i-- {
		hoverSelectors = append(hoverSelectors, fmt.Sprintf(
			".mj-carousel-%s-thumbnail:hover %s+ .mj-carousel-main .mj-carousel-image",
			carouselID,
			repeat(i),
		))
	}
	writeSelectorBlock(hoverSelectors, ",", "      display: none !important;")

	css.WriteString(fmt.Sprintf("    .mj-carousel-thumbnail:hover {\n      border-color: %s !important;\n    }\n\n", tbHoverBorderColor))

	showThumbSelectors := make([]string, imageCount)
	for i := 0; i < imageCount; i++ {
		showThumbSelectors[i] = fmt.Sprintf(
			".mj-carousel-%s-thumbnail-%d:hover %s+ .mj-carousel-main .mj-carousel-image-%d",
			carouselID,
			i+1,
			repeat(imageCount-i-1),
			i+1,
		)
	}
	css.WriteString("    ")
	css.WriteString(strings.Join(showThumbSelectors, ","))
	css.WriteString(" {\n")
	css.WriteString("      display: block !important;\n")
	css.WriteString("    }\n")
	css.WriteString("    \n\n")

	css.WriteString("      .mj-carousel noinput { display:block !important; }\n")
	css.WriteString("      .mj-carousel noinput .mj-carousel-image-1 { display: block !important;  }\n")
	css.WriteString("      .mj-carousel noinput .mj-carousel-arrows,\n")
	css.WriteString("      .mj-carousel noinput .mj-carousel-thumbnails { display: none !important; }\n\n")
	css.WriteString("      [owa] .mj-carousel-thumbnail { display: none !important; }\n")
	css.WriteString("      \n")
	css.WriteString("      @media screen yahoo {\n")
	css.WriteString(fmt.Sprintf("          .mj-carousel-%s-icons-cell,\n", carouselID))
	css.WriteString("          .mj-carousel-previous-icons,\n")
	css.WriteString("          .mj-carousel-next-icons {\n")
	css.WriteString("              display: none !important;\n")
	css.WriteString("          }\n\n")

	padding := " + *+"
	if imageCount > 2 {
		padding = " + *+ *+"
	}
	css.WriteString(fmt.Sprintf(
		"          .mj-carousel-%s-radio-1:checked%s .mj-carousel-content .mj-carousel-%s-thumbnail-1 {\n",
		carouselID,
		padding,
		carouselID,
	))
	css.WriteString("              border-color: transparent;\n")
	css.WriteString("          }\n")
	css.WriteString("      }")

	return css.String()
}

// generateCarouselID generates a unique ID for the carousel
func (c *MJCarouselComponent) generateCarouselID() string {
	if c.id != "" {
		return c.id
	}
	idValue := nextCarouselIDValue()
	c.id = fmt.Sprintf("%016x", idValue)
	return c.id
}

// nextCarouselIDValue returns the next pseudo-random ID in a deterministic sequence.
func nextCarouselIDValue() uint64 {
	for {
		current := atomic.LoadUint64(&carouselIDState)
		if current == 0 {
			if atomic.CompareAndSwapUint64(&carouselIDState, 0, carouselIDSeed) {
				return carouselIDSeed
			}
			continue
		}

		next := current*carouselIDMultiplier + carouselIDIncrement
		if atomic.CompareAndSwapUint64(&carouselIDState, current, next) {
			return next
		}
	}
}

// getCarouselImages gets all mj-carousel-image children
func (c *MJCarouselComponent) getCarouselImages() []*MJCarouselImageComponent {
	var images []*MJCarouselImageComponent
	for _, child := range c.Children {
		if carouselImage, ok := child.(*MJCarouselImageComponent); ok {
			images = append(images, carouselImage)
		}
	}
	return images
}

// renderCarouselContent renders the main carousel HTML content
func (c *MJCarouselComponent) renderCarouselContent(w io.StringWriter, carouselID string, carouselImages []*MJCarouselImageComponent) error {
	leftIcon := c.GetAttributeWithDefault(c, "left-icon")
	rightIcon := c.GetAttributeWithDefault(c, "right-icon")
	iconWidth := c.GetAttributeWithDefault(c, "icon-width")
	thumbnails := c.GetAttributeWithDefault(c, "thumbnails")

	// Start MSO conditional comment
	if _, err := w.WriteString("<!--[if !mso]><!-->"); err != nil {
		return err
	}

	// Main carousel container
	if _, err := w.WriteString(`<div class="mj-carousel">`); err != nil {
		return err
	}

	// Render radio buttons for carousel state management
	if err := c.renderRadioButtons(w, carouselID, len(carouselImages)); err != nil {
		return err
	}

	// Carousel content container
	if _, err := w.WriteString(`<div class="mj-carousel-content mj-carousel-`); err != nil {
		return err
	}
	if _, err := w.WriteString(carouselID); err != nil {
		return err
	}
	if _, err := w.WriteString(`-content" style="display:table;width:100%;table-layout:fixed;text-align:center;font-size:0px;">`); err != nil {
		return err
	}

	// Render thumbnails if visible
	if thumbnails == "visible" {
		if err := c.renderThumbnails(w, carouselID, carouselImages); err != nil {
			return err
		}
	}

	// Main carousel table with images and navigation
	if err := c.renderCarouselTable(w, carouselID, carouselImages, leftIcon, rightIcon, iconWidth); err != nil {
		return err
	}

	// Close carousel content and main container
	if _, err := w.WriteString(`</div></div>`); err != nil {
		return err
	}

	// Close MSO conditional comment
	if _, err := w.WriteString("<!--<![endif]-->"); err != nil {
		return err
	}

	return nil
}

// renderMSOFallback renders MSO-specific fallback content
func (c *MJCarouselComponent) renderMSOFallback(w io.StringWriter, carouselImages []*MJCarouselImageComponent) error {
	// MSO conditional comment for Outlook
	if _, err := w.WriteString("<!--[if mso]>"); err != nil {
		return err
	}

	// Show only first image in Outlook
	if len(carouselImages) > 0 {
		if err := c.renderCarouselImageContent(w, carouselImages[0], 1, "600", true); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("<![endif]-->"); err != nil {
		return err
	}

	return nil
}

// renderRadioButtons renders hidden radio buttons for carousel state management
func (c *MJCarouselComponent) renderRadioButtons(w io.StringWriter, carouselID string, imageCount int) error {
	for i := 1; i <= imageCount; i++ {
		// First radio button is checked by default
		checkedAttr := ""
		if i == 1 {
			checkedAttr = ` checked="checked"`
		}

		className := fmt.Sprintf("mj-carousel-radio mj-carousel-%s-radio mj-carousel-%s-radio-%d", carouselID, carouselID, i)

		if _, err := w.WriteString(fmt.Sprintf(`<input class="%s"%s type="radio" name="mj-carousel-radio-%s" id="mj-carousel-%s-radio-%d" style="display:none;mso-hide:all;">`,
			className, checkedAttr, carouselID, carouselID, i)); err != nil {
			return err
		}
	}
	return nil
}

// renderThumbnails renders thumbnail navigation images
func (c *MJCarouselComponent) renderThumbnails(w io.StringWriter, carouselID string, carouselImages []*MJCarouselImageComponent) error {
	tbBorder := c.GetAttributeWithDefault(c, "tb-border")
	tbBorderRadius := c.GetAttributeWithDefault(c, "tb-border-radius")
	tbWidth := c.GetAttributeWithDefault(c, "tb-width")

	for i, img := range carouselImages {
		imageNum := i + 1
		// Use thumbnails-src if available, otherwise fall back to src
		src := img.Node.GetAttribute("thumbnails-src")
		if src == "" {
			src = img.Node.GetAttribute("src")
		}
		href := fmt.Sprintf("#%d", imageNum)
		target := img.GetAttributeWithDefault(img, "target")

		// Build thumbnail CSS classes including any css-class from mj-carousel-image
		baseClasses := fmt.Sprintf("mj-carousel-thumbnail mj-carousel-%s-thumbnail mj-carousel-%s-thumbnail-%d", carouselID, carouselID, imageNum)
		imageClasses := img.Node.GetAttribute("css-class")
		if imageClasses != "" {
			baseClasses += " " + imageClasses + "-thumbnail"
		}

		// Thumbnail link
		if _, err := w.WriteString(fmt.Sprintf(`<a style="border:%s;border-radius:%s;display:inline-block;overflow:hidden;width:%s;" href="%s" target="%s" class="%s">`,
			tbBorder, tbBorderRadius, tbWidth, href, target, baseClasses)); err != nil {
			return err
		}

		// Thumbnail label and image
		alt := img.Node.GetAttribute("alt")
		altAttr := fmt.Sprintf(` alt="%s"`, alt)
		if _, err := w.WriteString(fmt.Sprintf(`<label for="mj-carousel-%s-radio-%d"><img style="display:block;width:100%%;height:auto;" src="%s"%s width="%s"></label>`,
			carouselID, imageNum, src, altAttr, strings.TrimSuffix(tbWidth, "px"))); err != nil {
			return err
		}

		if _, err := w.WriteString("</a>"); err != nil {
			return err
		}
	}
	return nil
}

// renderCarouselTable renders the main carousel table with images and navigation
func (c *MJCarouselComponent) renderCarouselTable(w io.StringWriter, carouselID string, carouselImages []*MJCarouselImageComponent, leftIcon, rightIcon, iconWidth string) error {
	// Main carousel table
	if _, err := w.WriteString(`<table border="0" cellpadding="0" cellspacing="0" role="presentation" width="100%" class="mj-carousel-main" style="caption-side:top;display:table-caption;table-layout:fixed;width:100%;"><tbody><tr>`); err != nil {
		return err
	}

	// Left navigation icons cell
	if _, err := w.WriteString(fmt.Sprintf(`<td class="mj-carousel-%s-icons-cell" style="font-size:0px;display:none;mso-hide:all;padding:0px;">`, carouselID)); err != nil {
		return err
	}

	if err := c.renderPreviousIcons(w, carouselID, len(carouselImages), leftIcon); err != nil {
		return err
	}

	if _, err := w.WriteString("</td>"); err != nil {
		return err
	}

	// Main images cell
	if _, err := w.WriteString(`<td style="padding:0px;"><div class="mj-carousel-images">`); err != nil {
		return err
	}

	// Render all carousel images
	for i, img := range carouselImages {
		imageNum := i + 1
		if err := c.renderCarouselImageContent(w, img, imageNum, "600", false); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("</div></td>"); err != nil {
		return err
	}

	// Right navigation icons cell
	if _, err := w.WriteString(fmt.Sprintf(`<td class="mj-carousel-%s-icons-cell" style="font-size:0px;display:none;mso-hide:all;padding:0px;">`, carouselID)); err != nil {
		return err
	}

	if err := c.renderNextIcons(w, carouselID, len(carouselImages), rightIcon); err != nil {
		return err
	}

	if _, err := w.WriteString("</td>"); err != nil {
		return err
	}

	// Close table
	if _, err := w.WriteString("</tr></tbody></table>"); err != nil {
		return err
	}

	return nil
}

// renderPreviousIcons renders the previous navigation icons
func (c *MJCarouselComponent) renderPreviousIcons(w io.StringWriter, carouselID string, imageCount int, leftIcon string) error {
	if _, err := w.WriteString(`<div class="mj-carousel-previous-icons" style="display:none;mso-hide:all;">`); err != nil {
		return err
	}

	iconWidth := c.GetAttributeWithDefault(c, "icon-width")

	for i := 1; i <= imageCount; i++ {
		iconWidthValue := strings.TrimSuffix(iconWidth, "px")
		if _, err := w.WriteString(fmt.Sprintf(`<label for="mj-carousel-%s-radio-%d" class="mj-carousel-previous mj-carousel-previous-%d"><img src="%s" alt="previous" style="display:block;width:%s;height:auto;" width="%s"></label>`,
			carouselID, i, i, leftIcon, iconWidth, iconWidthValue)); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("</div>"); err != nil {
		return err
	}
	return nil
}

// renderNextIcons renders the next navigation icons
func (c *MJCarouselComponent) renderNextIcons(w io.StringWriter, carouselID string, imageCount int, rightIcon string) error {
	if _, err := w.WriteString(`<div class="mj-carousel-next-icons" style="display:none;mso-hide:all;">`); err != nil {
		return err
	}

	iconWidth := c.GetAttributeWithDefault(c, "icon-width")

	for i := 1; i <= imageCount; i++ {
		iconWidthValue := strings.TrimSuffix(iconWidth, "px")
		if _, err := w.WriteString(fmt.Sprintf(`<label for="mj-carousel-%s-radio-%d" class="mj-carousel-next mj-carousel-next-%d"><img src="%s" alt="next" style="display:block;width:%s;height:auto;" width="%s"></label>`,
			carouselID, i, i, rightIcon, iconWidth, iconWidthValue)); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("</div>"); err != nil {
		return err
	}
	return nil
}

// renderCarouselImageContent renders a single carousel image
func (c *MJCarouselComponent) renderCarouselImageContent(w io.StringWriter, img *MJCarouselImageComponent, imageNum int, width string, isFallback bool) error {
	src := img.Node.GetAttribute("src")
	borderRadius := c.GetAttributeWithDefault(c, "border-radius")
	alt := img.Node.GetAttribute("alt")
	title := img.Node.GetAttribute("title")
	href := img.Node.GetAttribute("href")

	// Container div with CSS classes
	styleAttr := ""
	if isFallback {
		styleAttr = ` style="" `
	} else if imageNum > 1 {
		styleAttr = ` style="display:none;mso-hide:all;"`
	}

	// Build CSS classes for the container div
	containerClasses := fmt.Sprintf("mj-carousel-image mj-carousel-image-%d", imageNum)
	imageClasses := img.Node.GetAttribute("css-class")
	if imageClasses != "" {
		containerClasses += " " + imageClasses
	} else if isFallback {
		containerClasses += " "
	}

	if _, err := w.WriteString(fmt.Sprintf(`<div class="%s"%s>`, containerClasses, styleAttr)); err != nil {
		return err
	}

	// Add link wrapper if href is present
	if href != "" {
		if _, err := w.WriteString(fmt.Sprintf(`<a href="%s" target="_blank">`, href)); err != nil {
			return err
		}
	}

	// Image element with alt and title attributes
	altAttr := fmt.Sprintf(` alt="%s"`, alt)
	titleAttr := ""
	if title != "" {
		titleAttr = fmt.Sprintf(` title="%s"`, title)
	}
	var imgBuilder strings.Builder
	imgBuilder.WriteString("<img")
	if titleAttr != "" {
		imgBuilder.WriteString(titleAttr)
	}
	imgBuilder.WriteString(fmt.Sprintf(` src="%s"`, src))
	imgBuilder.WriteString(altAttr)
	imgBuilder.WriteString(fmt.Sprintf(` style="border-radius:%s;display:block;width:%spx;max-width:100%%;height:auto;"`, borderRadius, width))
	imgBuilder.WriteString(fmt.Sprintf(` width="%s"`, width))
	if isFallback {
		imgBuilder.WriteString(` border="0" />`)
	} else {
		imgBuilder.WriteString(` border="0">`)
	}

	if _, err := w.WriteString(imgBuilder.String()); err != nil {
		return err
	}

	// Close link wrapper if href was present
	if href != "" {
		if _, err := w.WriteString("</a>"); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("</div>"); err != nil {
		return err
	}

	return nil
}
