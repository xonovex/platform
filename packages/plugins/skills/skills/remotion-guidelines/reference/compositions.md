# filename: compositions

**Guideline:** Define compositions in `Root.tsx` with `type` declarations (not `interface`) and Zod `schema` for visual parameter UI; use `calculateMetadata()` for dynamic values.

**Rationale:** `type` works better with `defaultProps` and Zod inference; schema enables GUI parameter editing; `calculateMetadata` enables dynamic duration/dimensions from props.

**Example:**
```tsx
const schema = z.object({
  title: z.string(),
  backgroundColor: zColor(),
  fontSize: z.number().min(10).max(200)
});
type Props = z.infer<typeof schema>;

export function RemotionRoot() {
  return (
    <Composition<Props>
      id="MyVideo"
      component={MyVideo}
      durationInFrames={300}
      fps={30}
      width={1920}
      height={1080}
      schema={schema}
      defaultProps={{title: "Hello", backgroundColor: "#fff", fontSize: 48}}
    />
  );
}
```

**Techniques:**
- Composition setup: `id`, `component`, `durationInFrames`, `fps`, `width`, `height` required
- Type vs interface: Always use `type`, not `interface`; works better with Zod inference
- Schema UI: `z.object({...})` with `zColor()`, `z.number().min().max()`, `z.enum()` for visual editor
- Dynamic metadata: `calculateMetadata({props}) => ({durationInFrames: ...})` for async resolution
- Folder organization: Wrap `<Composition>` in `<Folder name="...">` for grouping
- Still images: `<Still>` for single frames; no `durationInFrames`/`fps` needed
- Props inference: Use `z.infer<typeof schema>` for type-safe props
- defaultProps required: Set sensible defaults for all schema fields
