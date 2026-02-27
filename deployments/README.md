# InkForge Deployment

## Docker Deployment

InkForge can be deployed using Docker containers for easy management and scaling.

### Building the Docker Image

```bash
# Build and run using docker compose
docker-compose up -d

# Or build manually
docker build -t inkforge .
```

### Running with Docker

```bash
# Run on default port (8080)
docker run -p 8080:8080 inkforge

# Run with custom port
docker run -p 3000:8080 -e PORT=8080 inkforge
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

### Production Considerations

- Use reverse proxy like nginx for SSL termination
- Monitor resource usage as Playwright can consume considerable memory
- Consider horizontal scaling for high-demand applications
- Regular security updates to base images