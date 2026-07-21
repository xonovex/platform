# Build stage
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.26.5-alpine3.24@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS build
ARG TARGETARCH
WORKDIR /src

# Copy shared dependencies first (for layer caching)
COPY packages/shared/shared-core-go/go.mod packages/shared/shared-core-go/go.sum* packages/shared/shared-core-go/
COPY packages/shared/shared-agent-go/go.mod packages/shared/shared-agent-go/go.sum* packages/shared/shared-agent-go/

# Copy operator module and download dependencies
COPY packages/agent/agent-operator-go/go.mod packages/agent/agent-operator-go/go.sum packages/agent/agent-operator-go/
WORKDIR /src/packages/agent/agent-operator-go
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
WORKDIR /src
COPY packages/shared/shared-core-go/ packages/shared/shared-core-go/
COPY packages/shared/shared-agent-go/ packages/shared/shared-agent-go/
COPY packages/agent/agent-operator-go/ packages/agent/agent-operator-go/

# Build static binary
WORKDIR /src/packages/agent/agent-operator-go
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /operator ./cmd/operator/

# Runtime stage
FROM gcr.io/distroless/static:nonroot@sha256:f7f8f729987ad0fdf6b05eeeae94b26e6a0f613bdf46feea7fc40f7bd72953e6 AS runtime
COPY --from=build --chown=65532:65532 /operator /operator
USER 65532:65532
EXPOSE 8081
ENTRYPOINT ["/operator"]
