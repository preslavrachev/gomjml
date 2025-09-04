# Container Width Flow in MJML Components

## Overview

MJML components use a cascading width calculation system where each parent component calculates the effective width available to its children after accounting for its own padding, borders, and margins.

## The Width Flow Chain

### 1. Section → Column
- **Section** provides its effective width (usually 600px) to columns
- **Column** receives parent width and calculates its content width by subtracting its own padding/borders
- **Column** sets this calculated width as `containerWidth` for all its children

### 2. Column → Child Components (Divider, Text, etc.)
- **Child components** receive `containerWidth` from column context
- **Child components** subtract their own padding to calculate final rendered width

## Example Calculation

Given this MJML structure:
```xml
<mj-section>  <!-- 600px width -->
  <mj-column mj-class="padded">  <!-- padding="20px" = 20px left + 20px right -->
    <mj-divider />  <!-- default padding="10px 25px" = 25px left + 25px right -->
  </mj-column>
</mj-section>
```

**Expected width flow:**
1. Section: 600px
2. Column effective width: 600px - 40px (column padding) = 560px
3. Column sets `containerWidth = 560px` for children
4. Divider width: 560px - 50px (divider padding) = **510px**

## MJML.io Reference Implementation

From the official MJML codebase (`packages/mjml-column/src/index.js` and `packages/mjml-divider/src/index.js`):

### Column Context (mjml-column/src/index.js)
```javascript
getChildContext() {
  // ... calculate allPaddings (left + right padding/borders) ...
  if (unit === '%') {
    containerWidth = `${
      (parseFloat(parentWidth) * parsedWidth) / 100 - allPaddings
    }px`
  } else {
    containerWidth = `${parsedWidth - allPaddings}px`
  }
  return {
    ...this.context,
    containerWidth,
  }
}
```

### Divider Width Calculation (mjml-divider/src/index.js)
```javascript
getOutlookWidth() {
  const { containerWidth } = this.context
  const paddingSize =
    this.getShorthandAttrValue('padding', 'left') +
    this.getShorthandAttrValue('padding', 'right')
  // ...
  return `${parseInt(containerWidth, 10) - paddingSize}px`
}
```

## Common Bug Pattern

**Symptom:** Divider renders at 550px instead of expected 510px (40px difference)

**Root Cause:** Column component is not properly subtracting its own padding before setting `containerWidth` for children.

**Fix:** Ensure column calculates effective content width and calls `child.SetContainerWidth(effectiveWidth)` for all children.

## Implementation Checklist

- [ ] Section passes its effective width to columns
- [ ] Column calculates: `effectiveWidth = receivedWidth - leftPadding - rightPadding - leftBorder - rightBorder`
- [ ] Column calls `child.SetContainerWidth(effectiveWidth)` for each child
- [ ] Child components use `GetContainerWidth()` and subtract their own padding for final rendering
- [ ] MSO tables in components use calculated width, not hardcoded values

## Debug Tips

1. **Add debug logging** to track width flow: Section → Column → Child
2. **Check MSO table widths** in generated HTML - they should match calculated widths
3. **Test with different padding combinations** to verify calculations
4. **Compare against MJML.io output** for the same template

## Detailed MJML.io Implementation Notes

### Column Width Calculation Details
- **Location**: `packages/mjml-column/src/index.js` in `getChildContext()` method
- **Box Model**: Uses `getBoxWidths()` helper to calculate total padding (left + right) and borders
- **Calculation**: 
  - Percentage width: `(parentWidth * percent) / 100 - allPaddings`
  - Pixel width: `parsedWidth - allPaddings`
- **Context**: Passes computed `containerWidth` as pixel string to children via React context

### Divider Width Usage Details  
- **Location**: `packages/mjml-divider/src/index.js` in `getOutlookWidth()` method
- **Process**:
  1. Gets `containerWidth` from parent context
  2. Calculates own left + right paddings
  3. Subtracts paddings from container width for effective width
  4. Applies percentage if width attribute specified
- **MSO Tables**: Uses `getOutlookWidth()` result for Outlook-specific table rendering

### Key Insight
The column **must** subtract paddings/borders before setting container width for children. The divider then subtracts its own paddings from the received container width.

## Related Files

- `mjml/components/section.go` - Section width calculation
- `mjml/components/column.go` - **[BUG LOCATION]** Column width calculation and context setting
- `mjml/components/divider.go` - Divider MSO width calculation  
- `mjml/components/base.go` - Container width interface methods

## AIDEV Notes

- AIDEV-NOTE: width-calc-bug; column must subtract paddings before SetContainerWidth() calls
- AIDEV-NOTE: divider-550px-hardcode; hardcoded 550px in MSO table should use calculated width
- AIDEV-TODO: implement getBoxWidths() helper for consistent padding/border calculations