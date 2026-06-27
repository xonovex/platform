# dpi-scaling: Own Per-Monitor DPI Scaling

## Guideline

The application owns DPI scaling, not the OS. Lay out the UI in virtual coordinates (1:1 at 100% / 96 DPI), apply the per-monitor scale at the edges (vertex shader and rect conversion), and key the font atlas by DPI so text stays sharp.

## Rationale

Letting the OS bitmap-rescale a high-DPI window blurs everything. Owning the scale keeps the immediate-mode core DPI-unaware (it works in virtual coordinates) while vector primitives, text, thumbnails, and viewports each scale correctly. Per-monitor awareness is essential because a window can straddle or move between monitors with different DPIs.

## How to Apply

1. Base mapping: 100% = 96 DPI; `dpi_scale = dpi_x / 96.0` (125% → 120 DPI, 150% → 144 DPI).
2. Call `SetProcessDpiAwareness(PROCESS_PER_MONITOR_DPI_AWARE)` **before creating any window**. Query monitor DPI with `GetDpiForMonitor(..., MDT_EFFECTIVE_DPI, ...)`; react to `WM_DPICHANGED` (new scale = `HIWORD(wparam) / 96.0`); pick a window's monitor by its **center pixel**.
3. Keep the IMGUI core in **virtual coordinates**; convert to pixels at the OS boundary (`adjust_rect(..., TO_PIXELS | TO_VIRTUAL)`) and apply the scale in the **vertex shader** so the core stays unscaled.
4. Generate font atlases on demand from TrueType (e.g. stb_truetype) at `pixels = ceil(points * dpi / 72.0)` with oversampling; **cache key = font + point size + monitor DPI**.
5. Thumbnails: size in actual pixels (so they stay sharp). 3D viewports: default to virtual-resolution rendering for battery/perf, make native resolution opt-in; keep back-buffer resolution distinct from render-target resolution.

## Example

```c
static float font_points_to_pixels(float pt, float dpi) { return ceilf(pt * dpi / 72.f); }
// atlas cache key must include dpi, or 96-DPI glyphs get stretched on a 144-DPI monitor
glyph_atlas_t *a = atlas_get(font_name, point_size, monitor_dpi);
```

## Counter-Example

A fixed-resolution kiosk or console UI on a single known display doesn't need per-monitor machinery — a single compile-time scale is enough. Per-monitor handling is for desktop apps spanning mixed-DPI displays.

## Gotcha

`SetThreadDpiAwarenessContext()` is thread-local, not global, and behaves surprisingly — set process awareness up front and don't rely on per-thread context for global state.

## Related

[draw-batching.md](./draw-batching.md)
