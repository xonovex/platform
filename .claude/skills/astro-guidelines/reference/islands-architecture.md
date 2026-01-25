# islands-architecture: Islands Architecture and Hydration

**Guideline:** Ship minimal JavaScript by hydrating only interactive components; static HTML by default.

**Rationale:** Astro renders to static HTML by default (zero JS). Framework components only hydrate when marked with directives, resulting in faster loads and better performance.

**Example:**

```astro
---
import Counter from "../components/Counter.tsx";
import SearchBar from "../components/SearchBar.tsx";
---
<Layout title="Home">
  <h1>Static Content</h1>
  <Counter client:visible />        <!-- Hydrate on viewport entry -->
  <SearchBar client:load />         <!-- Hydrate immediately -->
</Layout>
```

**Techniques:**
- Static by default: Build pages with Astro components that render zero JavaScript
- Framework components: Import React, Vue, Svelte only for interactive features
- client:load: Hydrate immediately for critical interactive components
- client:visible: Hydrate when component enters viewport for lazy interactivity
- client:idle: Hydrate during browser idle time for lower-priority features
