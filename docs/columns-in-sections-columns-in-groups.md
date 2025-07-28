# Columns in Sections vs. Columns in Groups

In MJML, the key difference between columns in a section versus columns in a group relates to **responsive behavior** and **layout control**:

## Columns in a Section (`mj-section`)

```xml
<mj-section>
  <mj-column>Content 1</mj-column>
  <mj-column>Content 2</mj-column>
</mj-section>
```

- **Responsive stacking**: Columns automatically stack vertically on mobile devices
- **Full-width container**: The section spans the entire email width
- **Automatic mobile optimization**: MJML handles the responsive breakpoints for you
- **Standard email layout**: This is the most common pattern for email layouts

## Columns in a Group (`mj-group`)

```xml
<mj-section>
  <mj-group>
    <mj-column>Content 1</mj-column>
    <mj-column>Content 2</mj-column>
  </mj-group>
</mj-section>
```

- **Prevents stacking**: Columns inside a group will **not** stack on mobile - they maintain their side-by-side layout
- **Nested within sections**: Groups must be placed inside sections
- **Fixed layout**: Useful when you need columns to stay horizontal even on small screens
- **Manual control**: You're responsible for ensuring the content fits properly on mobile

## When to Use Each

**Use columns in sections** (most common):
- Standard responsive email layouts
- When you want mobile-friendly stacking behavior
- For main content areas that should adapt to screen size

**Use groups**:
- When you need columns to stay side-by-side on mobile
- For small elements like social icons or buttons
- When you want precise control over mobile layout
- For creating complex multi-column layouts within a section

The group approach gives you more control but requires more careful consideration of how content will appear on smaller screens.