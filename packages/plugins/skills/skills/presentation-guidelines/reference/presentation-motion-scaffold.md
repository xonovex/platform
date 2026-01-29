# presentation-motion-scaffold: Create Motion Presentation Package

**Guideline:** Scaffold Motion React presentation package from markdown presentation document.

**Rationale:** Converts markdown presentations into fully-functional React + Motion slide decks with animations and interactive features.

**Example:**

```
Input: presentation.md with:
- Title: "Microservices Architecture"
- Theme colors: #2563eb, #dc2626, #16a34a
- Slides: overview, services, challenges, next steps

Output structure:
packages/presentations/microservices/
├── src/
│   ├── slides/
│   │   ├── Title.tsx
│   │   ├── Overview.tsx
│   │   ├── Services.tsx
│   │   └── NextSteps.tsx
│   ├── App.tsx
│   └── theme.ts (generated colors)
├── public/assets/
├── package.json (generated)
└── vite.config.ts (generated)
```

**Techniques:**

- Parse markdown: extract title, style guide, slides separated by `---`
- Determine slide types from content: title, bullets, diagrams, chapter headers, closing
- Create package structure at `packages/presentations/[name]/` with configs
- Generate TSX components for each slide type with factory functions
- Extract theme colors from Style Guide section (#hex format)
- Convert ASCII diagrams to GraphViz .dot files for SVG generation
- Download and place external logos/images to `public/assets/`
- Generate configs: package.json, moon.yml, tsconfig.json, vite.config.ts
- Create factory functions for each slide with automatic step calculation
