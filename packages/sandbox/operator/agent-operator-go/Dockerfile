# Build stage
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.27.4-alpine3.24@sha256:70bd7ddea8b6cbe5d6943cde279d6b67feb10cc924fb9100d1d8757c1575b280 AS build
ARG TARGETARCH
WORKDIR /src

# Copy shared dependencies first (for layer caching)
COPY packages/library/shared-core-go/go.mod packages/library/shared-core-go/go.sum* packages/library/shared-core-go/
COPY packages/library/shared-agent-go/go.mod packages/library/shared-agent-go/go.sum* packages/library/shared-agent-go/

# Copy operator module and download dependencies
COPY packages/sandbox/operator/agent-operator-go/go.mod packages/sandbox/operator/agent-operator-go/go.sum packages/sandbox/operator/agent-operator-go/
WORKDIR /src/packages/sandbox/operator/agent-operator-go
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy source
WORKDIR /src
COPY packages/library/shared-core-go/ packages/library/shared-core-go/
COPY packages/library/shared-agent-go/ packages/library/shared-agent-go/
COPY packages/sandbox/operator/agent-operator-go/ packages/sandbox/operator/agent-operator-go/

# Build static binary
WORKDIR /src/packages/sandbox/operator/agent-operator-go
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o /operator ./cmd/operator/

# Runtime stage
FROM gcr.io/distroless/static:nonroot@sha256:aafe993665bf6200db38a02c11a9351e782a84b6a45e7f33b6695638c88e3931 AS runtime
COPY --from=build --chown=65532:65532 /operator /operator
USER 65532:65532
EXPOSE 8081
ENTRYPOINT ["/operator"]
