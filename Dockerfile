# Build stage
FROM golang:1.25-alpine AS builder

# Build arguments for version injection
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG GIT_TREE_STATE=unknown
ARG BUILD_DATE=unknown

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.mod
COPY go.sum go.sum

# Cache dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY internal/ internal/

# TEMPLATE NOTE: Update the ldflags module path and binary name
# Change 'github.com/example-org/example-service' to your module path
# Change 'example-service' output name to your service name
# Change './cmd/example-service' to your cmd directory
# Build the binary with version information
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-X 'github.com/example-org/example-service/internal/version.Version=${VERSION}' \
              -X 'github.com/example-org/example-service/internal/version.GitCommit=${GIT_COMMIT}' \
              -X 'github.com/example-org/example-service/internal/version.GitTreeState=${GIT_TREE_STATE}' \
              -X 'github.com/example-org/example-service/internal/version.BuildDate=${BUILD_DATE}'" \
    -a -o example-service ./cmd/example-service

# Runtime stage
FROM gcr.io/distroless/static:nonroot

WORKDIR /
# TEMPLATE NOTE: Update binary name to match your service
COPY --from=builder /workspace/example-service .
USER 65532:65532

# TEMPLATE NOTE: Update entrypoint to match your binary name
ENTRYPOINT ["/example-service"]
