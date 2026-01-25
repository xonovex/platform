# textures: Texture Loading and Configuration

**Guideline:** Set correct color space per texture type (sRGB for colors, no color space for data), use power-of-2 dimensions, enable mipmaps.

**Rationale:** Color space errors cause washed-out/dark rendering; mipmaps ensure quality at distance; compression saves VRAM.

**Example:**
```javascript
const colorTex = loader.load('albedo.jpg');
colorTex.colorSpace = THREE.SRGBColorSpace;
const normalTex = loader.load('normal.jpg');
normalTex.colorSpace = THREE.NoColorSpace; // Data, not color
```

**Techniques:**
- Color spaces: sRGB for color/albedo/emissive, NoColorSpace for normal/roughness/metalness/displacement/AO
- Mipmaps: Auto-generate for most textures; disable for canvas/video (dynamic updates)
- Wrapping: RepeatWrapping or MirroredRepeatWrapping for tiling; ClampToEdgeWrapping default
- Filtering: LinearMipmapLinearFilter (smooth, quality), NearestFilter (pixelated retro)
- Anisotropy: `.anisotropy = renderer.capabilities.getMaxAnisotropy()` for sharp distant angles
- Compression: KTX2 (best modern), WebP (browser support), Data Textures (procedural)
