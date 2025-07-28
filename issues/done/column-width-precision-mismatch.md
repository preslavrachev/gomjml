# Column Width Precision Mismatch Issue

**Date:** 2025-07-28  
**Status:** Open  
**Priority:** High  
**Component:** Column width calculation (mjml/components/column.go)

## Summary

The gomjml implementation generates incorrect CSS width values for column classes compared to MRML reference implementation. The issue stems from improper floating-point precision handling when converting percentage widths to CSS class names.

## Problem Description

When MJML sections contain multiple columns, each column gets an automatic width calculated as `100% / column_count`. These widths are then used to generate CSS class names in the format `mj-column-per-{width}`. 

**Root Cause:** gomjml is truncating decimal precision in CSS width values, while MRML preserves full float32 precision.

## Detailed Findings

### MRML (Reference Implementation) Output
```css
.mj-column-per-10 { width:10% !important; max-width:10%; }
.mj-column-per-100 { width:100% !important; max-width:100%; }
.mj-column-per-11-111111 { width:11.111111% !important; max-width:11.111111%; }
.mj-column-per-12-5 { width:12.5% !important; max-width:12.5%; }
.mj-column-per-14-285714 { width:14.285714% !important; max-width:14.285714%; }
.mj-column-per-16-666666 { width:16.666666% !important; max-width:16.666666%; }
.mj-column-per-20 { width:20% !important; max-width:20%; }
.mj-column-per-25 { width:25% !important; max-width:25%; }
.mj-column-per-33-333332 { width:33.333332% !important; max-width:33.333332%; }
.mj-column-per-50 { width:50% !important; max-width:50%; }
```

### gomjml (Current Implementation) Output
```css
.mj-column-per-10 { width:10% !important; max-width:10%; }
.mj-column-per-100 { width:100% !important; max-width:100%; }
.mj-column-per-11-111111 { width:11% !important; max-width:11%; }
.mj-column-per-12-5 { width:12% !important; max-width:12%; }
.mj-column-per-14-285714 { width:14% !important; max-width:14%; }
.mj-column-per-16-666666 { width:17% !important; max-width:17%; }
.mj-column-per-20 { width:20% !important; max-width:20%; }
.mj-column-per-25 { width:25% !important; max-width:25%; }
.mj-column-per-33-333332 { width:33% !important; max-width:33%; }
.mj-column-per-50 { width:50% !important; max-width:50%; }
```

## Specific Discrepancies

| Columns | Expected Width | MRML Output | gomjml Output | Issue |
|---------|----------------|-------------|---------------|-------|
| 3 | 33.333332% | `width:33.333332%` | `width:33%` | Lost precision |
| 6 | 16.666666% | `width:16.666666%` | `width:17%` | Rounded up |
| 7 | 14.285714% | `width:14.285714%` | `width:14%` | Lost precision |
| 8 | 12.5% | `width:12.5%` | `width:12%` | Lost precision |
| 9 | 11.111111% | `width:11.111111%` | `width:11%` | Lost precision |

## Impact

1. **Visual Layout Differences:** Columns may appear narrower/wider than expected
2. **Integration Test Failures:** CSS comparisons fail due to precision mismatches
3. **Email Client Compatibility:** Different rendering in email clients due to width variations
4. **MRML Compatibility:** Cannot achieve identical output to reference implementation

## Technical Root Cause

Based on the MJML file comment, MRML uses **Rust's f32 (32-bit float) precision** with `%g` formatting. The gomjml implementation appears to be either:

1. Using integer conversion instead of proper float formatting
2. Using insufficient decimal precision in `fmt.Sprintf`
3. Applying premature rounding in width calculations

## Test Case

The issue is reproducible with the `column-width-test.mjml` file which tests automatic column width calculation for 1-10 columns per section.

```mjml
<!-- Example 3-column section -->
<mj-section>
  <mj-column><mj-text>3 Col A</mj-text></mj-column>
  <mj-column><mj-text>3 Col B</mj-text></mj-column>
  <mj-column><mj-text>3 Col C</mj-text></mj-column>
</mj-section>
```

**Expected:** `width:33.333332%`  
**Actual:** `width:33%`

## Proposed Solution

1. **Use float32 precision:** Convert width calculations to `float32` to match MRML
2. **Apply proper formatting:** Use Go's `%g` format verb or equivalent to match MRML's output
3. **Update GetColumnWidth function:** Ensure `mjml/components/column.go:GetColumnWidth` returns properly formatted percentages
4. **Preserve class name precision:** Maintain full decimal precision in CSS class names

## Files to Investigate

- `mjml/components/column.go` - Column width calculation logic
- `mjml/components/section.go` - Section component that calculates column widths
- `mjml/mediaquery/` - CSS media query generation

## Success Criteria

- All column width CSS values match MRML output exactly
- Integration tests pass for multi-column layouts
- `column-width-test.mjml` produces identical output to MRML

## Additional Context

This issue is critical for MRML compatibility and blocks several integration tests. The precision mismatch affects not just visual rendering but also email client compatibility where exact percentage values may influence layout behavior.