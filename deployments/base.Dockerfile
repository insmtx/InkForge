# InkForge base runtime environment
FROM registry.yygu.cn/library/playwright:v1.40.0-focal

# Install certificates and core utilities
RUN apt-get update && \
    apt-get install -y ca-certificates curl wget xz-utils git && \
    rm -rf /var/lib/apt/lists/*

# Create non-root user for security
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge

# Temporarily install Go to prepare Go Playwright assets
RUN wget -qO- https://golang.org/dl/go1.24.0.linux-amd64.tar.gz | tar xzf - -C /usr/local

# Set up Go environment temporarily to run playwright install
ENV GOROOT=/usr/local/go
ENV PATH=${GOROOT}/bin:${PATH}

# Create a temporary workspace for preparing playwright installation
RUN mkdir -p /tmp/playwright-prep && cd /tmp/playwright-prep && \
    go mod init playwright-prep && \
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
    echo '}' >> main.go && \
    go run main.go && \
    rm -rf /tmp/playwright-prep

# Clean up Go installation (we don't need it in the final image but browsers remain cached)
RUN rm -rf /usr/local/go

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