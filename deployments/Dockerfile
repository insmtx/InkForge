# InkForge release image with build process
# Multi-stage build to compile the application and create the final production image

# Stage 1: Build the Go application
FROM registry.yygu.cn/library/golang:1.24-bookworm AS builder

ENV GOPROXY=https://goproxy.cn,direct

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the InkForge binary with production flags
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s -extldflags '-static'" -installsuffix cgo -o inkforge ./cmd/inkforge

# Stage 2: Create final runtime image from the base image
FROM registry.yygu.cn/skills/inkforge:base

# Copy the compiled binary from the builder stage
COPY --from=builder --chown=inkforge:inkforge /workspace/inkforge /app/inkforge

# Set production health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Start the application
CMD ["./inkforge"]