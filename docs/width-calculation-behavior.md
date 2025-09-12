If a section is inside a wrapper with padding (e.g., <mj-wrapper padding="20px">), should the section's max-width be reduced by the wrapper's horizontal padding? If a section has its own padding (e.g., <mj-section padding="15px">), should this affect the section's max-width, or is this just internal padding within the section? In the specific case of:

```mjml
<mj-wrapper padding="20px">
<mj-section>...</mj-section>
</mj-wrapper>
```
Should the section have max-width: 560px (600-40) or max-width: 600px?

1. Yes. A section inside a wrapper inherits containerWidth from the wrapper’s getChildContext(), which is based on the wrapper’s box width (wrapper nominal width minus horizontal padding/borders). Wrapper horizontal padding therefore reduces the max-width available to its child sections so the overall outer width (content + padding) stays at 600px.

2. A section’s own padding does not change its max-width; the section still uses the passed containerWidth for its outer div (style max-width: <containerWidth>). That padding is internal spacing inside that max-width and further reduces the inner content (columns) box, but not the section’s declared max-width itself.

3. In your example:
<mj-wrapper padding="20px">
  <mj-section>...</mj-section>
</mj-wrapper>
The wrapper has 20px left + 20px right padding (40px total). The child section receives a containerWidth of 600 - 40 = 560, so the section should have max-width: 560px, not 600px.

Summary:
1. Reduce by wrapper horizontal padding: Yes.
2. Section padding affects inner content only, not its max-width value.
3. Resulting section max-width: 560px.