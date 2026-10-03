# components: Component Development and Composition

Build static UI as `.astro` components; reach for framework components only for interactivity. Beyond the default `<slot />`, expose named slots with fallback content:

```astro
<slot name="footer">{href && <a href={href}>Learn more →</a>}</slot>
```
