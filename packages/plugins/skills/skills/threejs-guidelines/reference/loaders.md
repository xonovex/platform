# filename: loaders

**Guideline:** Use GLTFLoader with Draco compression for 3D models; wrap loaders in async/await; use LoadingManager for progress tracking.

**Rationale:** GLTF/GLB is web standard; Draco reduces file size 90%+; async prevents UI blocking; manager centralizes progress.

**Example:**

```javascript
async function loadModel(url) {
  return new Promise((resolve, reject) => {
    new GLTFLoader().load(url, resolve, undefined, reject);
  });
}
const gltf = await loadModel('model.glb');
scene.add(gltf.scene);

// With Draco compression
const dracoLoader = new DRACOLoader();
dracoLoader.setDecoderPath('https://www.gstatic.com/draco/v1.5.6/');
new GLTFLoader().setDRACOLoader(dracoLoader).load('model.glb', ...);
```

**Techniques:**

- GLTFLoader: Primary format; animations in `gltf.animations`; use `traverse()` for shadow/material setup
- Draco: Reduce geometry size 90%+; requires decoder path setup
- KTX2 textures: Compressed texture format via `ktx2Loader.setTranscoderPath()`
- LoadingManager: Centralize `onProgress`/`onLoad`/`onError` across multiple loaders
- Parallel loading: `Promise.all([loadModel(), loadModel()])` for concurrent assets
- Export best practices: Enable Draco/KTX2 in 3D tools; use `gltf-transform` CLI for compression
