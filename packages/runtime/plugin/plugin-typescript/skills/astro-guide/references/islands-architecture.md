# islands-architecture: Islands Architecture and Hydration

Astro components render to static HTML with zero JS. Framework components ship no JavaScript until marked with a `client:*` directive: pick the cheapest one that works.

```astro
---
import Counter from "../components/Counter.tsx";
import SearchBar from "../components/SearchBar.tsx";
---

<Layout title="Home">
  <h1>Static Content</h1>
  <Counter client:visible />
  <!-- hydrate on viewport entry -->
  <SearchBar client:load />
  <!-- hydrate immediately -->
</Layout>
```

- `client:load`: hydrate immediately (critical interactivity)
- `client:idle`: hydrate on browser idle (lower priority)
- `client:media={query}`: hydrate when the media query matches
