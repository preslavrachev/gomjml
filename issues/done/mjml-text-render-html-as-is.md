# MJML Go Implementation Spec: Handling `<br>` Tags in `<mj-text>`

## Overview

This document specifies how the Go implementation of MJML should handle `<br>` (line break) tags—especially open, non-self-closing, or malformed variants—when present inside `<mj-text>` components. The behavior is based on the existing JavaScript MJML reference implementation and its documented semantics.

---

## 1. `mj-text` as an "Ending Tag" Component

- `<mj-text>` is an **ending tag** component.
- **Definition**: An "ending tag" component in MJML is one that can contain arbitrary HTML and text, but **cannot** contain other MJML components.
- All content inside `<mj-text>` is **passed through directly** to the output HTML without any additional MJML parsing, escaping, or correction.

**Reference (JS MJML):**
```markdown
`mj-text` is an "ending tag", which means it can contain HTML code which will be left as it is, so it can contain HTML tags with attributes, but it cannot contain other MJML components.
```

---

## 2. Behavior for `<br>` Tags

### 2.1 Valid HTML `<br>` Tags

- Both `<br>` and `<br/>` inside `<mj-text>` are legal HTML and are **preserved as-is** in the output HTML.

**Example:**
```xml
<mj-text>
  foo<br>bar
</mj-text>
```
**Renders as:**
```html
<div>foo<br>bar</div>
```

### 2.2 Open or Malformed `<br>` Tags

- Any HTML that is not well-formed (e.g., `<br bar`) is also passed through unaltered, as long as the MJML document itself is valid XML.
- If the malformed tag causes the MJML to become invalid XML, the MJML parser (in Go, as in JS) should reject the document with a parsing error.

**Example:**
```xml
<mj-text>
  foo<br bar
</mj-text>
```
**Renders as:**
```html
<div>foo<br bar</div>
```
- Rendering is then the responsibility of the email client or browser.

### 2.3 No Correction or Auto-Closing

- The MJML Go implementation **must not attempt to fix or close** HTML tags inside `<mj-text>`.
- What you put inside `<mj-text>` is what you get in the output HTML.

---

## 3. Requirement: Input MJML Must Be Valid XML

- The MJML document must be valid XML for the parser to accept it.
- This means that `<mj-text>` itself must be properly closed.
- However, HTML fragments **inside `<mj-text>` do not need to be valid XML or HTML**.

---

## 4. Examples

### Example 1: Standard Usage

```xml
<mj-text>
  Hello<br>World!
</mj-text>
```
**Output:**
```html
<div>Hello<br>World!</div>
```

### Example 2: Self-Closing HTML Tag

```xml
<mj-text>
  Hello<br/>World!
</mj-text>
```
**Output:**
```html
<div>Hello<br />World!</div>
```
**Note:** XML parsers normalize self-closing tags by adding a space before `/>`, so `<br/>` becomes `<br />`.

### Example 3: Malformed HTML Tag

```xml
<mj-text>
  Hello<br bar>World!
</mj-text>
```
**Output:**
```html
<div>Hello<br bar>World!</div>
```
- The output HTML is responsible for rendering or ignoring malformed tags.

### Example 4: Invalid XML (Should Fail)

```xml
<mj-text>
  Hello<br>World!
<!-- missing closing </mj-text> -->
```
**Result:**  
Parser error: MJML document is invalid XML.

---

## 5. Justification and References

- All HTML within `<mj-text>` is passed through unchanged, per the JS MJML implementation:
  - [doc/ending-tags.md](https://github.com/mjmlio/mjml/blob/main/doc/ending-tags.md)
  - [mjml-text/README.md](https://github.com/mjmlio/mjml/blob/main/packages/mjml-text/README.md)
  - [mjml-text/src/index.js](https://github.com/mjmlio/mjml/blob/main/packages/mjml-text/src/index.js)
- The Go implementation should mirror this behavior for compatibility.
- The responsibility for rendering (or ignoring) malformed HTML tags is left to the browser/email client.

---

## 6. Summary Table

| Scenario                        | Input in `<mj-text>`    | Output HTML (inside `<div>`) | Notes                        |
| ------------------------------- | ----------------------- | ---------------------------- | ---------------------------- |
| Valid `<br>`                    | `foo<br>bar`            | `foo<br>bar`                 | Pass through as-is           |
| Self-closing `<br/>`            | `foo<br/>bar`           | `foo<br />bar`               | XML parser normalizes with space |
| Self-closing `<br />`           | `foo<br />bar`          | `foo<br />bar`               | Already has space, preserved |
| Malformed tag (e.g. `<br bar>`) | `foo<br bar>bar`        | `foo<br bar>bar`             | Pass through as-is           |
| Missing closing `</mj-text>`    | `foo<br>bar` (no close) | Parser error                 | MJML input must be valid XML |

---

## 7. Implementation Notes (for Go Developers)

- MJML Go parser should treat `<mj-text>` as an "ending tag" and only process its attributes.
- The content of `<mj-text>` is copied verbatim to the output HTML.
- **No sanitization, escaping, or parsing** of HTML inside `<mj-text>` should occur.
- This rule applies to all HTML tags, not just `<br>` and `<img>`.

### XML Parser Normalization Behavior

- **Self-closing tags**: XML parsers normalize `<tag/>` to `<tag />` (adding space before `/>`)
- **Void elements**: Both `<br>`, `<br/>`, and `<br />` are valid HTML and should be supported
- **Other void elements**: Same behavior applies to `<img>`, `<hr>`, `<input>`, etc.
- **Malformed HTML**: Should pass through unchanged (e.g., `<br bar>`, unclosed `<strong>`)

### Current Implementation Gaps

1. **Unclosed void elements** like `<br>` and `<img>` fail XML parsing
2. **Malformed HTML** fails XML parsing  
3. **Unclosed non-void elements** like `<strong>` fail XML parsing
4. **Self-closing normalization** works correctly (adds space)

The challenge is handling HTML that isn't valid XML while preserving the XML structure of the MJML document itself.
