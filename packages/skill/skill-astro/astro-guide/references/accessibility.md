# accessibility: Accessibility Best Practices

**accessibility-guide** owns applicable criteria, assessment evidence, exceptions, and conformance claims. This reference owns the Astro implementation delta: semantic template markup and required `alt` text on Astro's `<Image>`.

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
