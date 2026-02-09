# Build stage
FROM golang:1.25-alpine AS build
WORKDIR /src

# Copy shared dependencies first (for layer caching)
COPY packages/shared/shared-core-go/go.mod packages/shared/shared-core-go/go.sum* packages/shared/shared-core-go/
COPY packages/shared/shared-agent-go/go.mod packages/shared/shared-agent-go/go.sum* packages/shared/shared-agent-go/

# Copy operator module and download dependencies
COPY packages/agent/agent-operator-go/go.mod packages/agent/agent-operator-go/go.sum packages/agent/agent-operator-go/
WORKDIR /src/packages/agent/agent-operator-go
RUN go mod download

# Copy source
WORKDIR /src
COPY packages/shared/shared-core-go/ packages/shared/shared-core-go/
COPY packages/shared/shared-agent-go/ packages/shared/shared-agent-go/
COPY packages/agent/agent-operator-go/ packages/agent/agent-operator-go/

# Build static binary
WORKDIR /src/packages/agent/agent-operator-go
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /operator ./cmd/operator/

# Runtime stage
FROM gcr.io/distroless/static:nonroot AS runtime
COPY --from=build --chown=65532:65532 /operator /operator
USER 65532:65532
EXPOSE 8081
ENTRYPOINT ["/operator"]
