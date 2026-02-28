# InkForge base runtime environment
FROM registry.yygu.cn/library/golang:1.24 AS installer

ENV GOPROXY=https://goproxy.cn,direct
# Install dependencies needed for playwright installation
RUN apt-get update && \
    apt-get install -y ca-certificates curl git && \  
    rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code needed for install command
COPY ./ ./

# Build the InkForge binary specifically to run the install command
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-w -s -extldflags '-static'" -installsuffix cgo -o inkforge ./cmd/inkforge

# Execute the install command to prepare Playwright browsers and dependencies
RUN ./inkforge install

# Final stage - Create base image with all pre-installed dependencies
FROM registry.yygu.cn/library/playwright:v1.40.0-focal

# Install certificates and core utilities needed at runtime
RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge

# Copy pre-installed Playwright cache from installer stage
COPY --from=installer --chown=inkforge:inkforge /root/.cache /root/.cache

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
