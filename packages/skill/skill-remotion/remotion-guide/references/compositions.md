# compositions: Composition Structure and Metadata

Register `<Composition>` in `Root.tsx` with required `id`, `component`, `durationInFrames`, `fps`, `width`, `height`. Type props with `type` + `z.infer` (not `interface` — better `defaultProps`/Zod inference); pass a Zod `schema` (`zColor()`, `z.number().min().max()`, `z.enum()`) for the visual param editor and `defaultProps` for every field. Resolve dynamic duration/dimensions async via `calculateMetadata({props}) => ({durationInFrames, ...})`.

```tsx
const schema = z.object({
  title: z.string(),
  backgroundColor: zColor(),
  fontSize: z.number().min(10).max(200),
});
type Props = z.infer<typeof schema>;

<Composition<Props>
  id="MyVideo"
  component={MyVideo}
  durationInFrames={300}
  fps={30}
  width={1920}
  height={1080}
  schema={schema}
  defaultProps={{title: "Hello", backgroundColor: "#fff", fontSize: 48}}
/>;
```

- `<Still>` for single frames (no `durationInFrames`/`fps`); wrap in `<Folder name="...">` to group in the sidebar
