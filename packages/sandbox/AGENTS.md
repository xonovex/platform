# Sandbox

Keep Kubernetes execution policy in the operator and image construction in the image package.

- The operator owns Kubernetes execution policy. The image package builds and publishes that operator from the repository root so its shared Go libraries remain in the Docker build context.
