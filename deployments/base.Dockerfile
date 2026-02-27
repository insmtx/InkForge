# InkForge base runtime environment
FROM registry.yygu.cn/library/playwright:v1.40.0-focal

# Install certificates and core utilities
RUN apt-get update && \
    apt-get install -y ca-certificates curl && \
    rm -rf /var/lib/apt/lists/*

# Install required browsers for Playwright 
RUN npx playwright install chromium

# Create non-root user for security
RUN groupadd -r inkforge && useradd -r -g inkforge inkforge

# Define working directory for the application
WORKDIR /app

# Expose default port
EXPOSE 8080

# Health check placeholder - to be overridden by services
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD echo "Health check not configured"