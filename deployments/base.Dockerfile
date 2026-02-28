# InkForge base runtime environment
FROM golang:1.24-bookworm AS playwright-builder

# Install nodejs and other dependencies for playwright installation
RUN apt-get update && \
    apt-get install -y curl git ca-certificates xz-utils && \
    rm -rf /var/lib/apt/lists/*

# Install Playwright dependencies in the builder stage  
WORKDIR /tmp/playwright-build
RUN go mod init playwright-build && \
    go mod edit -require=github.com/playwright-community/playwright-go@latest && \
    go mod tidy && \
    echo 'package main' > main.go && \
    echo 'import "fmt"' >> main.go && \
    echo 'import "github.com/playwright-community/playwright-go"' >> main.go && \
    echo 'func main() {' >> main.go && \
    echo '  err := playwright.Install()' >> main.go && \
    echo '  if err != nil {' >> main.go && \
    echo '    fmt.Printf("Error installing Playwright: %v\n", err)' >> main.go && \
    echo '  } else {' >> main.go && \
    echo '    fmt.Println("Successfully installed Playwright")' >> main.go && \
    echo '  }' >> main.go && \
    echo '}' >> main.go

# Build and run to pre-install the drivers/browsers
RUN go run main.go

# Final stage: copy the pre-installed Playwright cache to the base image
FROM registry.yygu.cn/library/playwright:v1.40.0-focal

# Install certificates and core utilities
RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge

# Copy pre-installed Playwright cache from builder stage
COPY --from=playwright-builder --chown=root:root /root/.cache /root/.cache

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

# Create non-root user for security
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge

# Define working directory for the application
WORKDIR /app

# Expose default port
EXPOSE 8080

# Health check placeholder - to be overridden by services
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD echo "Health check not configured"