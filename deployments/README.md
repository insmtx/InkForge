# InkForge Deployment

## Docker Deployment

InkForge can be deployed using Docker containers for easy management and scaling.

The deployment consists of:
- Base environment: `deployments/base.Dockerfile` - Contains the minimal Playwright environment
- Application image: `Dockerfile` - Contains the build and runtime for InkForge

### Building the Images

```bash
# Build the base environment image (optional, for custom base)
docker build -f deployments/base.Dockerfile -t inkforge-base .

# Build and run using docker compose (builds main image)
docker-compose up -d

# Or build main image manually
docker build -t inkforge .
```

### Running with Docker

```bash
# Run on default port (8080)
docker run -p 8080:8080 inkforge

# Run with custom port
docker run -p 3000:8080 -e PORT=8080 inkforge

# Run with custom environment variables
docker run -p 8080:8080 \
  -e PORT=8080 \
  -e KATEX_ENABLED=true \
  -e MERMAID_ENABLED=true \
  -e MAX_CONTENT_LENGTH=2097152 \
  inkforge
```

### Environment Variables

The following environment variables can be configured:

- `PORT` - Port to bind to (default: 8080)
- `HOST` - Host to bind to (default: 0.0.0.0)
- `KATEX_ENABLED` - Enable LaTex mathematics rendering (default: true)
- `MERMAID_ENABLED` - Enable Mermaid diagram rendering (default: true)
- `MAX_CONTENT_LENGTH` - Maximum Markdown content length in bytes (default: 1048576)
- `READ_TIMEOUT` - HTTP read timeout in seconds (default: 30)
- `WRITE_TIMEOUT` - HTTP write timeout in seconds (default: 60)

See the [docker-compose.yml](../docker-compose.yml) file for configuration examples.

### Health Check

The application provides a health check endpoint at `/api/v1/health` or `/health`.

### Base Environment

The `deployments/base.Dockerfile` provides a minimal, reusable Playwright environment with:
- Playwright installed with Chromium browser
- Minimal OS packages
- Security-hardened non-root user
- Optimized for headless rendering tasks

### Production Considerations

- Use reverse proxy like nginx for SSL termination
- Monitor resource usage as Playwright can consume considerable memory
- Consider horizontal scaling for high-demand applications
- Regular security updates to base images
- Consider using multi-stage builds in CI/CD pipelines