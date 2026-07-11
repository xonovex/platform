# components: Component Development and Composition

Build static UI as `.astro` components: type props via an `interface Props` destructured from `Astro.props`, compose with `<slot>` (default and `name`d), and scope CSS with `<style>`. Reach for framework components only for interactivity.

```astro
---
interface Props {
  title: string;
  href?: string;
  variant?: "default" | "featured";
}
const {title, href, variant = "default"} = Astro.props;
---

<div class={`card card--${variant}`}>
  <h3>{title}</h3>
  <slot />
  <slot name="footer">{href && <a href={href}>Learn more →</a>}</slot>
</div>
<style>
  .card--featured {
    border-color: #0066cc;
  }
</style>
```
