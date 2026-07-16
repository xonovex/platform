# textures: Texture Loading and Configuration

Set `NoColorSpace` on data textures (normal/roughness/metalness/AO) — they encode vectors, not color; use power-of-2 dimensions and enable mipmaps.

```javascript
const normalTex = loader.load("normal.jpg");
normalTex.colorSpace = THREE.NoColorSpace; // Data, not color
```

## Techniques

- Color spaces: `NoColorSpace` for data maps (normal/roughness/metalness/displacement/AO)
- Mipmaps: Auto-generate for most textures; disable for canvas/video (dynamic updates)
- Wrapping: RepeatWrapping or MirroredRepeatWrapping for tiling; ClampToEdgeWrapping default
- Filtering: LinearMipmapLinearFilter (smooth, quality), NearestFilter (pixelated retro)
- Anisotropy: `.anisotropy = renderer.capabilities.getMaxAnisotropy()` for sharp distant angles
- Compression: KTX2 (best modern), WebP (browser support), Data Textures (procedural)
