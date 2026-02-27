# Multi-stage build for InkForge
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

# Stage 2: Runtime environment with minimal footprint
FROM mcr.microsoft.com/playwright:v1.40.0-focal

# Install core dependencies only
RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Copy the binary from the builder stage
COPY --from=builder /app/inkforge .

# Ensure Playwright browsers are installed
RUN playwright install chromium

# Create non-root user
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge
WORKDIR /home/inkforge
COPY --from=builder --chown=inkforge:inkforge /app/inkforge .

# Switch to non-root user
USER inkforge

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

# Run the binary
CMD ["./inkforge"]