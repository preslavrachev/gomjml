# Group vs Section Column Comparison Analysis

## Executive Summary

**Key Finding**: The issues are **NOT group-specific** but affect **both group and section column rendering**. However, the problems manifest differently:

- **mj-group**: **Structural mismatch** with major table count differences (expected 5, actual 3)
- **mj-section with columns**: **Minor formatting differences** with same DOM structure but content variations

## Test Results Comparison

### mj-group.mjml Test Results
- **Status**: ❌ **FAILING** (Major structural issues)
- **Table Count**: Expected 5 tables, gomjml generates 3 
- **Issue Type**: DOM structure mismatch with extensive CSS property placement errors
- **Severity**: **HIGH** - Completely different HTML structure

### mj-section-with-columns.mjml Test Results  
- **Status**: ✅ **PASSING** (Fixed - 2025-07-28)
- **Table Count**: ✅ **MATCHES** (Same DOM structure)
- **Issue Type**: Resolved after shared fixes
- **Severity**: **RESOLVED** - MSO width and image height format issues fixed

## Structural Analysis

### Key Differences Between Group vs Section

#### 1. MRML Expected Structure Differences

**Group Structure** (with group wrapper):
```html
<div class="mj-column-per-100 mj-outlook-group-fix">  <!-- Group wrapper -->
  <!--[if mso | IE]><table><tr><![endif]-->
  <!--[if mso | IE]><td style="width:300px;"><![endif]-->
  <div class="mj-outlook-group-fix mj-column-per-50">  <!-- Column -->
    <table style="vertical-align:top;">               <!-- Column table -->
      <!-- Column content -->
    </table>
  </div>
  <!--[if mso | IE]></td></tr></table><![endif]-->
</div>
```

**Section Structure** (direct columns):
```html
<!--[if mso | IE]><table><tr><![endif]-->
<!--[if mso | IE]><td style="width:300px;"><![endif]-->
<div class="mj-outlook-group-fix mj-column-per-50">    <!-- Direct column -->
  <table style="vertical-align:top;">                 <!-- Column table -->
    <!-- Column content -->
  </table>
</div>
<!--[if mso | IE]></td></tr></table><![endif]-->
```

#### 2. CSS Class Generation Differences

**Group MRML** includes additional CSS:
```css
.mj-column-per-100 { width:100% !important; max-width:100%; }
```

**Section MRML** omits this class (only generates column-specific classes):
```css
.mj-column-per-50 { width:50% !important; max-width:50%; }
```

## Common Issues Affecting Both

### 1. MSO Width Calculation Bug
**Both tests show identical MSO width error:**

| Expected | Actual | Component |
|----------|--------|-----------|
| `width:300px` | `width:150px` | Both group and section columns |

This indicates a **shared bug in `GetWidthAsPixel()` method** used by both components.

### 2. Image Height Attribute Format
**Both tests show minor formatting difference:**

| Expected | Actual | Issue |
|----------|--------|-------|
| `height="185"` | `height="185px"` | Unit inconsistency |

### 3. Text Content Formatting
**Both tests have minor whitespace differences** in paragraph content, suggesting **shared text rendering logic**.

## Root Cause Analysis

### Group-Specific Issues (Major)
1. **Missing Group Wrapper**: gomjml doesn't generate the outer `mj-column-per-100` wrapper div
2. **Incorrect Table Nesting**: Group implementation doesn't create the proper 2-level table structure
3. **MSO Conditional Logic**: Group columns need different MSO rendering than section columns

### Shared Issues (Affecting Both)
1. **MSO Width Calculation**: `GetWidthAsPixel()` returns 150px instead of 300px for 50% columns
2. **Image Attribute Format**: Height format inconsistency (with/without "px")
3. **Text Content Processing**: Minor whitespace/formatting differences in text rendering

## Impact Assessment

### mj-group Impact (High Priority)
- **Broken email layout** in Outlook/MSO clients
- **Responsive behavior failures** due to missing group wrapper
- **Complete visual differences** from expected MJML output

### mj-section Impact (RESOLVED ✅)  
- **Fixed**: MSO width calculation now correct (300px vs 150px)
- **Fixed**: Image height format now matches MRML (no "px" suffix)  
- **Status**: All section-with-columns tests now passing

## Fix Strategy

### Phase 1: Fix Shared Issues (COMPLETED ✅)
1. **✅ Fixed `GetWidthAsPixel()` calculation** in column component
   - Location: `mjml/components/column.go:171`
   - Solution: Use `BaseComponent.GetEffectiveWidth()` instead of column's own width
   - Result: ✅ **Fixed both group and section MSO width** (300px instead of 150px)

2. **✅ Standardized image height format**
   - Location: `mjml/components/image.go:66-68,127`
   - Solution: Strip "px" suffix from height attribute like width
   - Result: ✅ **Fixed both group and section image rendering** (height="185" not "185px")

**Phase 1 Results**: 
- ✅ **mj-section-with-columns test now PASSES completely**
- ⚠️ **mj-group test still has structural issues** (table count mismatch, missing wrapper)

### Phase 2: Fix Group-Specific Issues (Group Only)
1. **Implement group wrapper div** (`mj-column-per-100 mj-outlook-group-fix`)
2. **Add proper MSO table nesting** for group columns
3. **Update CSS generation** to include group-specific classes

## Testing Strategy

### Verification Order
1. **Fix shared issues first** - validate both tests improve
2. **Test section-with-columns** - should **PASS** after shared fixes
3. **Fix group-specific issues** - focus on structural problems
4. **Test mj-group** - should **PASS** after group fixes

### Test Commands
```bash
# Test both after shared fixes
go test ./mjml -run TestMJMLAgainstMRML/mj-section-with-columns -v
go test ./mjml -run TestMJMLAgainstMRML/mj-group -v

# Verify improvements
diff -u /tmp/mrml_section_columns.html /tmp/gomjml_section_columns.html
diff -u /tmp/mrml_mj_group.html /tmp/gomjml_mj_group.html
```

## Technical Debt Assessment

### Column Component Architecture
- **Shared Logic Issues**: Both group and section columns use same base logic, but group needs special handling
- **Width Calculation Bug**: Fundamental error affecting multiple components
- **MSO Rendering**: Need unified approach to MSO conditional rendering

### Group Component Implementation  
- **Incomplete Implementation**: Missing critical wrapper and nesting logic
- **Integration Issues**: Poor integration with column component's group context
- **CSS Generation**: Group-specific styles not properly generated

## Conclusion

**The issues are NOT purely group-specific** but represent **two distinct problem categories**:

1. **Shared Column Issues** (affecting both): MSO width calculation, image formatting, text processing
2. **Group Architecture Issues** (group-only): Missing wrapper structure, incomplete MSO nesting

**Recommendation**: Fix shared issues first to improve both tests, then address group-specific structural problems for complete mj-group compatibility.

---

**Created**: 2025-07-28  
**Updated**: 2025-07-28  
**Status**: Partial Resolution (Section: ✅ Complete, Group: ⚠️ Ongoing)  
**Priority**: High (Group remaining issues)  
**Category**: Component Architecture Analysis