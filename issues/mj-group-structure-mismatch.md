# mj-group Structure Mismatch Analysis

## Issue Summary

The `mj-group.mjml` test is failing due to **fundamental structural differences** between MRML's group implementation and gomjml's current implementation. The test shows **table element count mismatch** (expected 5, actual 3) and extensive CSS property misalignment.

## Test Analysis

### Test Output Summary
- **Status**: FAILING ❌
- **Table Count**: Expected 5 tables, gomjml generates 3
- **Primary Issues**: 
  - Missing MSO conditional table structure around group columns
  - CSS property placement on wrong HTML elements
  - `vertical-align:top` property missing from column tables

### Key Differences Identified

#### 1. MSO Conditional Structure Mismatch

**MRML Expected Structure** (from reference output):
```html
<div class="mj-column-per-100 mj-outlook-group-fix" style="...">
  <!--[if mso | IE]><table><tr><![endif]-->
  <!--[if mso | IE]><td style="vertical-align:top;width:300px;"><![endif]-->
  <div class="mj-outlook-group-fix mj-column-per-50" style="...">
    <table style="vertical-align:top;">  <!-- KEY: This table has vertical-align -->
      <tbody>
        <tr>
          <td style="font-size:0px;padding:0;word-break:break-word;">
            <table style="border-collapse:collapse;border-spacing:0px;">
              <!-- mj-image content -->
            </table>
          </td>
        </tr>
        <!-- mj-text content -->
      </tbody>
    </table>
  </div>
  <!--[if mso | IE]></td><![endif]-->
  <!-- Second column follows same pattern -->
  <!--[if mso | IE]></tr></table><![endif]-->
</div>
```

**gomjml Current Structure**:
```html
<div class="mj-column-per-100 mj-outlook-group-fix" style="...">
  <!--[if mso | IE]><table><tr><![endif]-->
  <!--[if mso | IE]><td style="vertical-align:top;width:150px;"><![endif]-->
  <div class="mj-outlook-group-fix mj-column-per-50" style="...">
    <table style="border-collapse:collapse;border-spacing:0px;"> <!-- MISSING vertical-align:top -->
      <!-- Direct column content without proper nesting -->
    </table>
  </div>
  <!--[if mso | IE]></td><![endif]-->
  <!--[if mso | IE]></tr></table><![endif]-->
</div>
```

#### 2. CSS Property Misplacement

The integration test reveals systematic CSS property misplacement:

| Element Type | Expected Properties | Actual Properties | Issue |
|--------------|-------------------|------------------|-------|
| `table` (column wrapper) | `vertical-align:top` | `border-collapse:collapse;border-spacing:0px` | Missing key property |
| `td` (column cell) | `font-size:0px;padding:0;word-break:break-word` | `width:137px` | Wrong properties entirely |
| `img` | `width:137px` | Full image styles | Property moved to wrong element |

#### 3. Table Nesting Structure

**MRML uses 2-level table nesting within columns:**
1. **Outer table**: Column wrapper with `vertical-align:top`
2. **Inner table**: Component-specific (image/text) with `border-collapse:collapse`

**gomjml currently uses single-level nesting** which collapses the structure.

## Root Cause Analysis

### 1. Group Column Rendering Logic (`mjml/components/group.go:103-142`)

The current implementation:
- Sets `InsideGroup = true` flag for child columns
- Relies on child columns to handle group-specific rendering
- Missing the intermediate table wrapper that MRML uses

### 2. Column Component Group Context (`mjml/components/column.go`)

The column component likely needs to:
- Render differently when `InsideGroup = true`
- Create the outer table wrapper with `vertical-align:top`
- Properly nest component tables inside

### 3. MSO Conditional Width Calculation

**Current Issue**:
```
<!--[if mso | IE]><td style="vertical-align:top;width:150px;"><![endif]-->
```

**Expected**:
```
<!--[if mso | IE]><td style="vertical-align:top;width:300px;"><![endif]-->
```

The MSO width calculation appears incorrect (150px vs 300px for 50% columns).

## Specific Failing Elements

### Element Analysis from Test Output

1. **Table element[6]**: Missing `vertical-align:top` on column table
2. **TD element[7]**: Properties assigned to wrong element type
3. **Image element[9]**: Width calculation wrong (`width:100%` vs `width:137px`)
4. **Missing elements[19-20]**: Expected 2 additional table/div elements not generated

## Fix Strategy

### Phase 1: Column Group Context Handling
1. **Update `MJColumnComponent.Render()`** to detect `InsideGroup = true`
2. **Add group-specific table wrapper** with `vertical-align:top` style
3. **Ensure proper CSS property placement** on correct HTML elements

### Phase 2: MSO Width Calculation Fix
1. **Fix MSO conditional width calculation** in `GetWidthAsPixel()` method
2. **Ensure 50% columns render as 300px** for 600px container
3. **Test against different column counts** (2, 3, 4 columns)

### Phase 3: CSS Property Verification
1. **Audit all CSS properties** against MRML reference for each element type
2. **Ensure insertion order matches** MRML (critical for test passing)
3. **Verify image width handling** within group context

## Implementation Priority

| Priority | Component | Action Required | Estimated Effort |
|----------|-----------|----------------|------------------|
| **HIGH** | `column.go` | Add group context rendering logic | 2-3 hours |
| **HIGH** | `group.go` | Fix MSO width calculation | 1 hour |
| **MEDIUM** | `image.go` | Group-specific width handling | 1 hour |
| **LOW** | Integration test | Verify all edge cases | 30 minutes |

## Test Verification

Once implemented, verify with:
```bash
# Primary test
go test ./mjml -run TestMJMLAgainstMRML/mj-group -v

# Related group tests
go test ./mjml -run TestMJMLAgainstMRML -v | grep -i group

# Full integration suite
go test ./mjml -run TestMJMLAgainstMRML -v
```

## Related Issues

This issue likely affects:
- Any MJML files using `mj-group` components
- Multi-column layouts within groups
- MSO/Outlook compatibility for grouped content
- Responsive design behavior on mobile devices

## Technical Debt Notes

- The current `InsideGroup` flag approach is correct but incomplete
- Group component needs tighter integration with column rendering
- MSO conditional rendering needs systematic audit across all components
- Consider extracting common MSO table generation logic to shared utility

---

**Created**: 2025-07-28  
**Status**: Open  
**Priority**: High  
**Category**: Component Implementation