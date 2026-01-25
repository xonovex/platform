# accessibility: Accessibility Best Practices

**Guideline:** Use semantic HTML, descriptive alt text, and ARIA attributes for accessibility.

**Rationale:** Static-first approach enables accessible sites by default; semantic HTML ensures assistive tech compatibility and improves SEO.

**Example:**

```astro
<header>
  <nav aria-label="Main navigation">
    <ul><li><a href="/">Home</a></li></ul>
  </nav>
</header>

<main>
  <article>
    <h1>Title</h1>
    <Image src={image} alt="Descriptive text" />
    <section><h2>Section</h2></section>
  </article>
  <aside aria-label="Related">...</aside>
</main>
```

**Techniques:**
- Semantic elements: Use `<header>`, `<nav>`, `<main>`, `<article>`, `<aside>` for structure
- Alt text: Provide descriptive alt attributes for all images
- Heading hierarchy: Maintain h1 → h2 → h3 logical ordering
- ARIA labels: Add `aria-label` to navigation and complex regions
- Navigation region: Use `<nav aria-label>` for clear landmark structure
