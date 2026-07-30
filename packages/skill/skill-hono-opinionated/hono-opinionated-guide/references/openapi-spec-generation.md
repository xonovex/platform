# openapi-spec-generation: Generate the Spec with app.doc()

Register the spec endpoint with `app.doc(path, metadata)` on an `OpenAPIHono` app: it builds `paths` from the registered routes. Never hand-maintain a static OpenAPI document or the `paths` object. Supply only the top-level metadata: `openapi` version, `info`, optional `servers`.

```typescript
app.doc("/openapi.json", {
  openapi: "3.1.0",
  info: {title: "Hono Backend API", version: "1.0.0"},
  servers: [{url: "http://localhost:3000", description: "Development"}],
});
```
