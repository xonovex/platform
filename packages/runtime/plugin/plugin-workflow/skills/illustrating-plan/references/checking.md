# Checking and re-rendering

Run these checks after every render. The render commands are in [rendering.md](rendering.md).

## Check the images

Open every image you changed and read it against the plan and the decision files:

- Each participant is a component the decisions name, and the state mark of each element matches its decision.
- The lead image holds no open element.
- The option image shows every option of its decision and no other.
- The labels fit, no note is clipped, and the image is not exactly 4096 pixels wide, which means it was cropped.
- Nothing sensitive is visible, as [labels.md](labels.md) defines it.

Use the harness's image viewer. For an SVG that the viewer cannot show, render a PNG of the same source into a scratch directory outside the bundle with `--png -o <scratch>` and look at that, then delete the scratch copy. A check of the source text is not a check of the image.

`ls` must show an image for every source and a source for every image. An image under another base name than its source means the render did not do what the source says: find out why before reporting.

## Re-rendering

After a decision changes, re-render only the sources the change touched. Then run `git status` on `diagrams/`: a source you did not touch should show no change in its image. If it does, the renderer version differs from the one that drew it. Say so in the reply and re-render every source with the one renderer and the one format, so the images agree. A source for an option or a decision that no longer exists is deleted together with its image.
