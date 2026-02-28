# InkForge base runtime environment
FROM golang:1.24-bookworm AS builder

# Install the needed dependencies for Go build
RUN apt-get update && \
    apt-get install -y ca-certificates curl git && \
    rm -rf /var/lib/apt/lists/*

# Set up workspace for building the inkforge binary with install command
WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the InkForge binary with install command support
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s -extldflags '-static'" -installsuffix cgo -o inkforge ./cmd/inkforge

# Builder stage for installing Playwright dependencies
FROM golang:1.24-bookworm AS playwright-installer

# Install dependencies needed for playwright installation
RUN apt-get update && \
    apt-get install -y ca-certificates curl && \  
    rm -rf /var/lib/apt/lists/*

WORKDIR /tmp/playwright-install

# Copy the built binary from the first stage 
COPY --from=builder /workspace/inkforge ./

# Run the install subcommand to prepare browsers and dependencies
RUN ./inkforge install

# Final stage: copy the pre-installed Playwright cache to create the base image
FROM registry.yygu.cn/library/playwright:v1.40.0-focal

# Install certificates and core utilities needed at runtime
RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge

# Copy pre-installed Playwright cache from installer stage
COPY --from=playwright-installer --chown=root:root /root/.cache /root/.cache

# Set up cache directory with proper permissions for the app user
RUN chown -R inkforge:inkforge /root/.cache && \
    chmod -R 755 /root/.cache

# Define working directory for the application
WORKDIR /app

# Expose default port
EXPOSE 8080

# Health check placeholder - to be overridden by services
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD echo "Health check not configured"