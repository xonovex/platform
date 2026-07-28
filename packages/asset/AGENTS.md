# Assets

- Asset packages are private and stay at version `0.0.0`; they carry no `npm` tag, so they sit outside the lockstep release line that the skill and command plugins share.
- Every asset package owns a `ci-check` that proves its committed binaries are what they claim, because nothing else reads them: `asset-images` asserts the PNG MIME type, `asset-diagrams` re-renders each PNG from its source. Both tasks glob the whole directory, so adding a file adds it to the gate; never name one file in a task.
- In `asset-diagrams` the `.dot` file is the source and the `.png` is a build output that is also committed, so a reviewer sees the rendered diagram. Edit the `.dot`, then run `npx moon run asset-diagrams:graph-build` and commit both.
- Rendering needs `dot`; `nix/general.nix` supplies graphviz to the dev shell and CI.
- A screenshot in `asset-images` has no source to regenerate it from, so treat it as a dated snapshot: it records whatever tool versions, model names, and clock its window chrome happened to show. Do not cite those values as current anywhere, and do not re-capture merely because a version number moved.
- Re-capture a screenshot when what it demonstrates stops being true, for example when the sandbox axes, the CLI invocation, or the harness layout changes. Keep the pane count and tiling stable so the documents embedding it keep their layout.
- Reference every image from a document. `README.md` link targets are checked by `script-moon-release-validate`, so an image linked there is verified to exist on every run; an unreferenced image is dead weight that only the MIME check touches.
