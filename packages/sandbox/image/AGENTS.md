# Images

Build operator images from the repository root so operator code and shared Go libraries remain in the build context.

- Keep operator images in the agent execution environment group. Read image sources from the operator package and shared code from `packages/library/`.
