# InkForge release image with build process
# Multi-stage build to compile the application and create the final production image

# Stage 1: Build the Go application
FROM registry.yygu.cn/library/golang:1.24 AS builder

ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.org
ENV GONOPROXY=""
ENV GONOSUMDB=""

WORKDIR /workspace

# Create a temporary volume mount point for module and build caches to leverage Docker cache
RUN mkdir -p /go/pkg /root/.cache/go-build

# Create a non-root user for build process (security best practice)
RUN groupadd -g 1001 appuser && useradd -u 1001 -g appuser appuser

# Cache go modules by copying go.mod and go.sum first
COPY go.mod go.sum ./

# Download dependencies with caching
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && \
    go mod verify

# Set up caching for go build process
ENV GOCACHE=/root/.cache/go-build
ENV GOMODCACHE=/go/pkg/mod

# Copy source code last, after go modules are cached
COPY . .

# Build the InkForge binary with production flags and leverage Go build cache
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -v -ldflags="-w -s -extldflags '-static'" -installsuffix cgo -o inkforge ./cmd/inkforge

# Stage 2: Create final runtime image from the base image
FROM registry.yygu.cn/skills/inkforge:base

COPY ./internal/render/template/base.html /app/template/base.html
COPY ./demo/ /app/demo/
# Copy the compiled binary from the builder stage
COPY --from=builder /workspace/inkforge /app/inkforge

# Set production health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Start the application
CMD ["./inkforge"]