package components

import (
	"fmt"
	"io"
	"strings"
	"sync/atomic"

	"github.com/preslavrachev/gomjml/mjml/options"
	"github.com/preslavrachev/gomjml/parser"
)

// Global counter for unique carousel IDs
var carouselIDCounter int64

// ResetCarouselIDCounter resets the global counter for deterministic testing
func ResetCarouselIDCounter() {
	atomic.StoreInt64(&carouselIDCounter, 0)
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
	if _, err := w.WriteString(` style="font-size:0px;word-break:break-word;">`); err != nil {
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

	buildLevelPadding := func(level int) string {
		padding := " +"
		for i := 0; i < level; i++ {
			padding += " * +"
		}
		return padding
	}

	// Base carousel styles
	css.WriteString(".mj-carousel { -webkit-user-select: none;\n")
	css.WriteString("-moz-user-select: none;\n")
	css.WriteString("user-select: none; }\n")

	// Icon cell styles
	css.WriteString(fmt.Sprintf(".mj-carousel-%s-icons-cell { display: table-cell !important;\n", carouselID))
	css.WriteString(fmt.Sprintf("width: %s !important; }\n", iconWidth))

	// Hide radio buttons and navigation by default
	css.WriteString(".mj-carousel-radio,\n")
	css.WriteString(".mj-carousel-next,\n")
	css.WriteString(".mj-carousel-previous { display: none !important; }\n")

	// Touch action for thumbnails and navigation
	css.WriteString(".mj-carousel-thumbnail,\n")
	css.WriteString(".mj-carousel-next,\n")
	css.WriteString(".mj-carousel-previous { touch-action: manipulation; }\n")

	// Hide all images by default when radio button is checked
	css.WriteString(fmt.Sprintf(".mj-carousel-%s-radio:checked%s .mj-carousel-content .mj-carousel-image", carouselID, buildLevelPadding(0)))
	if imageCount > 1 {
		css.WriteString(",\n")
		css.WriteString(fmt.Sprintf(".mj-carousel-%s-radio:checked%s .mj-carousel-content .mj-carousel-image", carouselID, buildLevelPadding(1)))
	}
	if imageCount > 2 {
		css.WriteString(",\n")
		css.WriteString(fmt.Sprintf(".mj-carousel-%s-radio:checked%s .mj-carousel-content .mj-carousel-image", carouselID, buildLevelPadding(2)))
	}
	css.WriteString(" { display: none !important; }\n")

	// Show specific images when their radio button is checked
	for i := 1; i <= imageCount; i++ {
		padding := buildLevelPadding(imageCount - i)
		css.WriteString(fmt.Sprintf(".mj-carousel-%s-radio-%d:checked%s .mj-carousel-content .mj-carousel-image-%d", carouselID, i, padding, i))
		if i < imageCount {
			css.WriteString(",\n")
		}
	}
	css.WriteString(" { display: block !important; }\n")

	// Navigation icons visibility
	css.WriteString(".mj-carousel-previous-icons,\n")
	css.WriteString(".mj-carousel-next-icons")
	for i := 1; i <= imageCount; i++ {
		nextImage := i%imageCount + 1
		prevImage := imageCount
		if i > 1 {
			prevImage = i - 1
		}
		padding := buildLevelPadding(imageCount - i)
		css.WriteString(fmt.Sprintf(",\n.mj-carousel-%s-radio-%d:checked%s .mj-carousel-content .mj-carousel-next-%d", carouselID, i, padding, nextImage))
		css.WriteString(fmt.Sprintf(",\n.mj-carousel-%s-radio-%d:checked%s .mj-carousel-content .mj-carousel-previous-%d", carouselID, i, padding, prevImage))
	}
	css.WriteString(" { display: block !important; }\n")

	// Thumbnail selection styles
	for i := 1; i <= imageCount; i++ {
		padding := buildLevelPadding(imageCount - i)
		css.WriteString(fmt.Sprintf(".mj-carousel-%s-radio-%d:checked%s .mj-carousel-content .mj-carousel-%s-thumbnail-%d", carouselID, i, padding, carouselID, i))
		if i < imageCount {
			css.WriteString(",\n")
		}
	}
	css.WriteString(fmt.Sprintf(" { border-color: %s !important; }\n", tbSelectedBorderColor))

	// Hide div after images
	css.WriteString(".mj-carousel-image img + div,\n")
	css.WriteString(".mj-carousel-thumbnail img + div { display: none !important; }\n")

	// Hover effects for thumbnails
	for k := imageCount - 1; k >= 1; k-- {
		css.WriteString(fmt.Sprintf(".mj-carousel-%s-thumbnail:hover%s .mj-carousel-main .mj-carousel-image,\n", carouselID, buildLevelPadding(k)))
	}
	css.WriteString(fmt.Sprintf(".mj-carousel-%s-thumbnail:hover + .mj-carousel-main .mj-carousel-image { display: none !important; }\n", carouselID))

	css.WriteString(fmt.Sprintf(".mj-carousel-thumbnail:hover { border-color: %s !important; }\n", tbHoverBorderColor))

	// Show image on thumbnail hover
	for i := 1; i <= imageCount; i++ {
		padding := buildLevelPadding(imageCount - i)
		css.WriteString(fmt.Sprintf(".mj-carousel-%s-thumbnail-%d:hover%s .mj-carousel-main .mj-carousel-image-%d", carouselID, i, padding, i))
		if i < imageCount {
			css.WriteString(",\n")
		}
	}
	css.WriteString(" { display: block !important; }\n")

	// Fallback styles for no input support
	css.WriteString(".mj-carousel noinput { display:block !important; }\n")
	css.WriteString(".mj-carousel noinput .mj-carousel-image-1 { display: block !important;  }\n")
	css.WriteString(".mj-carousel noinput .mj-carousel-arrows, .mj-carousel noinput .mj-carousel-thumbnails { display: none !important; }\n")

	// Outlook Web App thumbnail hiding
	css.WriteString("[owa] .mj-carousel-thumbnail { display: none !important; }\n")

	// Media query for screen and yahoo
	css.WriteString("\n        @media screen, yahoo {\n")
	css.WriteString(fmt.Sprintf("            .mj-carousel-%s-icons-cell,\n", carouselID))
	css.WriteString("            .mj-carousel-previous-icons,\n")
	css.WriteString("            .mj-carousel-next-icons {\n")
	css.WriteString("                display: none !important;\n")
	css.WriteString("            }\n\n")
	padding := " + *+"
	if imageCount > 2 {
		padding = " + *+ *+"
	}
	css.WriteString(fmt.Sprintf("            .mj-carousel-%s-radio-1:checked%s .mj-carousel-content .mj-carousel-%s-thumbnail-1 {\n", carouselID, padding, carouselID))
	css.WriteString("                border-color: transparent;\n")
	css.WriteString("            }\n")
	css.WriteString("        }")

	return css.String()
}

// generateCarouselID generates a unique ID for the carousel
func (c *MJCarouselComponent) generateCarouselID() string {
	if c.id != "" {
		return c.id
	}
	counter := atomic.LoadInt64(&carouselIDCounter)
	c.id = fmt.Sprintf("%08d", counter) // Start from 00000000
	atomic.AddInt64(&carouselIDCounter, 1)
	return c.id
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
		if err := c.renderCarouselImageContent(w, carouselImages[0], 1, "600"); err != nil {
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
		checked := ""
		if i == 1 {
			checked = ` checked="checked"`
		}

		if _, err := w.WriteString(fmt.Sprintf(`<input%s type="radio" name="mj-carousel-radio-%s" id="mj-carousel-%s-radio-%d" class="mj-carousel-radio mj-carousel-%s-radio mj-carousel-%s-radio-%d" style="display:none;mso-hide:all;" />`,
			checked, carouselID, carouselID, i, carouselID, carouselID, i)); err != nil {
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
		if _, err := w.WriteString(fmt.Sprintf(`<a href="%s" target="%s" class="%s" style="border:%s;border-radius:%s;display:inline-block;overflow:hidden;width:%s;">`,
			href, target, baseClasses, tbBorder, tbBorderRadius, tbWidth)); err != nil {
			return err
		}

		// Thumbnail label and image
		alt := img.Node.GetAttribute("alt")
		altAttr := ""
		if alt != "" {
			altAttr = fmt.Sprintf(` alt="%s"`, alt)
		}
		if _, err := w.WriteString(fmt.Sprintf(`<label for="mj-carousel-%s-radio-%d"><img src="%s"%s width="%s" style="display:block;width:100%%;height:auto;" /></label>`,
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
		if err := c.renderCarouselImageContent(w, img, imageNum, "600"); err != nil {
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
		if _, err := w.WriteString(fmt.Sprintf(`<label for="mj-carousel-%s-radio-%d" class="mj-carousel-previous mj-carousel-previous-%d"><img src="%s" alt="previous" width="%s" style="display:block;width:%s;height:auto;" /></label>`,
			carouselID, i, i, leftIcon, iconWidthValue, iconWidth)); err != nil {
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
		if _, err := w.WriteString(fmt.Sprintf(`<label for="mj-carousel-%s-radio-%d" class="mj-carousel-next mj-carousel-next-%d"><img src="%s" alt="next" width="%s" style="display:block;width:%s;height:auto;" /></label>`,
			carouselID, i, i, rightIcon, iconWidthValue, iconWidth)); err != nil {
			return err
		}
	}

	if _, err := w.WriteString("</div>"); err != nil {
		return err
	}
	return nil
}

// renderCarouselImageContent renders a single carousel image
func (c *MJCarouselComponent) renderCarouselImageContent(w io.StringWriter, img *MJCarouselImageComponent, imageNum int, width string) error {
	src := img.Node.GetAttribute("src")
	borderRadius := c.GetAttributeWithDefault(c, "border-radius")
	alt := img.Node.GetAttribute("alt")
	title := img.Node.GetAttribute("title")
	href := img.Node.GetAttribute("href")

	// Container div with CSS classes
	style := ""
	if imageNum > 1 {
		style = ` style="display:none;mso-hide:all;"`
	}

	// Build CSS classes for the container div
	containerClasses := fmt.Sprintf("mj-carousel-image mj-carousel-image-%d", imageNum)
	imageClasses := img.Node.GetAttribute("css-class")
	if imageClasses != "" {
		containerClasses += " " + imageClasses
	}

	if _, err := w.WriteString(fmt.Sprintf(`<div class="%s"%s>`, containerClasses, style)); err != nil {
		return err
	}

	// Add link wrapper if href is present
	if href != "" {
		if _, err := w.WriteString(fmt.Sprintf(`<a href="%s" target="_blank">`, href)); err != nil {
			return err
		}
	}

	// Image element with alt and title attributes
	altAttr := ""
	if alt != "" {
		altAttr = fmt.Sprintf(` alt="%s"`, alt)
	}
	titleAttr := ""
	if title != "" {
		titleAttr = fmt.Sprintf(` title="%s"`, title)
	}
	if _, err := w.WriteString(fmt.Sprintf(`<img border="0"%s src="%s"%s width="%s" style="border-radius:%s;display:block;width:%spx;max-width:100%%;height:auto;" />`,
		altAttr, src, titleAttr, width, borderRadius, width)); err != nil {
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
