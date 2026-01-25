# components: Component Development and Composition

**Guideline:** Build reusable Astro components with props and slots; use framework components only for interactivity.

**Rationale:** Astro components for static content, framework components (React, Vue, Svelte) for interactive features; minimizes JavaScript.

**Example:**

```astro
---
// Card.astro
interface Props {
  title: string;
  description?: string;
  href?: string;
  variant?: "default" | "featured";
}

const { title, description, href, variant = "default" } = Astro.props;
---
<div class={`card card--${variant}`}>
  <h3>{title}</h3>
  {description && <p>{description}</p>}
  <div class="card__content"><slot /></div>
  <div class="card__footer">
    <slot name="footer">
      {href && <a href={href}>Learn more →</a>}
    </slot>
  </div>
</div>
<style>
  .card { border: 1px solid #ddd; padding: 1rem; border-radius: 8px; }
  .card--featured { border-color: #0066cc; background: #f0f8ff; }
</style>
```

**Techniques:**
- Astro components: Create `.astro` components for static presentational UI
- Props typing: Use TypeScript interfaces for type-safe component props
- Slots composition: Leverage slots for flexible component composition
- Framework separation: Import framework components only for interactivity
- Scoped styles: Use `<style>` blocks for component-scoped CSS
