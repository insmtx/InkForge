# Multi-stage build for InkForge production image
# Stage 1: Build the application
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the InkForge binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o inkforge ./cmd/inkforge

# Stage 2: Runtime environment with Playwright
FROM mcr.microsoft.com/playwright:v1.40.0-focal

# Install certificates and core utilities
RUN apt-get update && \
    apt-get install -y ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# Install required browsers for Playwright
RUN playwright install chromium

# Create non-root user for security
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge

# Define working directory
WORKDIR /app

# Copy the compiled binary from builder stage
COPY --from=builder /app/inkforge .

# Switch to non-root user
USER inkforge

# Expose port
EXPOSE 8080

# Set health check for production use
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Run the binary
CMD ["./inkforge"]