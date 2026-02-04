# Deployment Guide

## Google Cloud Run Deployment

### Prerequisites
- Google Cloud account
- `gcloud` CLI installed and authenticated
- Docker (optional, Cloud Run can build for you)

### Quick Deploy (Recommended)

Cloud Run can build directly from source:

```bash
# Set your project
gcloud config set project YOUR_PROJECT_ID

# Deploy (Cloud Run will detect Dockerfile and build)
gcloud run deploy developer-events-mcp \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --memory 256Mi \
  --timeout 300 \
  --concurrency 100 \
  --min-instances 0 \
  --max-instances 10
```

### Manual Docker Build & Deploy

If you prefer to build locally:

```bash
# Build the container
docker build -t gcr.io/YOUR_PROJECT_ID/developer-events-mcp .

# Push to Google Container Registry
docker push gcr.io/YOUR_PROJECT_ID/developer-events-mcp

# Deploy
gcloud run deploy developer-events-mcp \
  --image gcr.io/YOUR_PROJECT_ID/developer-events-mcp \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --memory 256Mi
```

### Get Service URL

```bash
gcloud run services describe developer-events-mcp \
  --region us-central1 \
  --format 'value(status.url)'
```

### Configure Claude Desktop

Update `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "developer-events": {
      "url": "https://YOUR-SERVICE-URL.run.app"
    }
  }
}
```

### Cost Estimation

Cloud Run free tier includes:
- 2 million requests per month
- 360,000 GB-seconds of memory
- 180,000 vCPU-seconds

This MCP server should easily fit within free tier for personal use:
- Minimal memory (256MB)
- Fast response times (<1s)
- 1-hour cache reduces API calls
- Cold start: ~100-200ms

Expected cost for moderate use (100 requests/day): **$0.00/month**

## Other Cloud Platforms

### AWS ECS/Fargate

```bash
# Build and tag
docker build -t developer-events-mcp .

# Push to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com
docker tag developer-events-mcp:latest YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/developer-events-mcp:latest
docker push YOUR_ACCOUNT.dkr.ecr.us-east-1.amazonaws.com/developer-events-mcp:latest

# Create ECS service with Fargate
# (Use AWS Console or CloudFormation)
```

### Fly.io

```bash
# Install flyctl
curl -L https://fly.io/install.sh | sh

# Launch app
fly launch

# Deploy
fly deploy
```

### Railway

1. Connect your GitHub repository
2. Railway will auto-detect Dockerfile
3. Deploy automatically

### Azure Container Apps

```bash
# Create resource group
az group create --name mcp-rg --location eastus

# Create container app
az containerapp create \
  --name developer-events-mcp \
  --resource-group mcp-rg \
  --image YOUR_REGISTRY/developer-events-mcp:latest \
  --target-port 8080 \
  --ingress external \
  --query properties.configuration.ingress.fqdn
```

## Testing the Deployment

### Health Check

```bash
curl https://YOUR-SERVICE-URL/health
# Should return: OK
```

### MCP Connection Test

The MCP server uses streamable HTTP transport. Test with any MCP client or Claude Desktop.

### Monitor Logs

**Cloud Run:**
```bash
gcloud run services logs read developer-events-mcp --region us-central1
```

**Docker local:**
```bash
docker logs CONTAINER_ID
```

## Troubleshooting

### Container won't start
- Check that PORT environment variable is set (Cloud Run does this automatically)
- Verify the image architecture matches your platform
- Check logs for startup errors

### Claude Desktop can't connect
- Ensure the URL is accessible (try `curl https://YOUR-URL/health`)
- Check that `--allow-unauthenticated` is set in Cloud Run
- Verify the URL in Claude Desktop config is correct (no trailing slash)

### High latency
- First request will be slower due to cache warming (~500ms-1s)
- Subsequent requests use 1-hour cache and are very fast (<100ms)
- Consider setting `--min-instances 1` to avoid cold starts (costs ~$5/month)

### Memory issues
- Default 256Mi should be plenty
- Check actual usage: `gcloud run services describe developer-events-mcp --region us-central1`
- Increase if needed, but typically stays under 50MB

## Security Considerations

### Public vs Private

**Public (recommended for testing):**
- `--allow-unauthenticated` - anyone can use the MCP server
- Good for public service
- No authentication needed in Claude Desktop config

**Private (recommended for production):**
```bash
gcloud run deploy developer-events-mcp \
  --no-allow-unauthenticated
```

Then configure authentication:
```json
{
  "mcpServers": {
    "developer-events": {
      "url": "https://YOUR-SERVICE-URL.run.app",
      "headers": {
        "Authorization": "Bearer YOUR_TOKEN"
      }
    }
  }
}
```

### Rate Limiting

Consider adding rate limiting if running publicly:
- Use Cloud Run's built-in concurrency limits
- Add API Gateway in front
- Implement rate limiting in the code

### API Key Protection

The developers.events API is public and doesn't require keys. If you modify this to use private APIs:
- Store API keys in Google Secret Manager
- Pass as environment variables to Cloud Run
- Never commit keys to git
