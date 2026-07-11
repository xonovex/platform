# accessibility: Accessibility Best Practices

Use semantic landmarks (`<header>`, `<nav>`, `<main>`, `<article>`, `<aside>`), a logical h1→h2→h3 heading order, `aria-label` on nav and complex regions, and always pass `alt` to Astro's `<Image>`.

```astro
<nav aria-label="Main navigation">
  <ul><li><a href="/">Home</a></li></ul>
</nav>
<main>
  <article>
    <h1>Title</h1>
    <Image src={image} alt="Descriptive text" />
  </article>
  <aside aria-label="Related">...</aside>
</main>
```
