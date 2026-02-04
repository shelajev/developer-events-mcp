# Developer.events MCP Server 🎯

**Single binary MCP server with ZERO dependencies!**

A native MCP (Model Context Protocol) server built with Go that provides access to developer conference CFPs from [developers.events](https://developers.events/).

## ⚡ Why This?

- ✅ **True native binary** - no runtime required!
- ✅ **Instant startup** - milliseconds, not seconds
- ✅ **Tiny memory footprint** - ~10-20MB
- ✅ **Single file distribution** - just download and run
- ✅ **Cross-platform** - Linux, macOS (Intel & Apple Silicon), Windows

## 📦 Download & Run

### Quick Install

**Linux (x86_64):**
```bash
curl -L https://github.com/shelajev/developer-events-mcp/releases/latest/download/developer-events-mcp-linux-amd64 -o mcp-server
chmod +x mcp-server
./mcp-server
```

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/shelajev/developer-events-mcp/releases/latest/download/developer-events-mcp-darwin-arm64 -o mcp-server
chmod +x mcp-server
./mcp-server
```

**macOS (Intel):**
```bash
curl -L https://github.com/shelajev/developer-events-mcp/releases/latest/download/developer-events-mcp-darwin-amd64 -o mcp-server
chmod +x mcp-server
./mcp-server
```

**Windows:**
Download `developer-events-mcp-windows-amd64.exe` from [releases](https://github.com/shelajev/developer-events-mcp/releases) and run it.

## 🚀 Quick Start

This server can run in **two modes**:

### Mode 1: Local (stdio) - For Claude Desktop

1. **Configure Claude Desktop**

Add to your Claude Desktop config:

**macOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "developer-events": {
      "command": "/full/path/to/mcp-server"
    }
  }
}
```

2. **Restart Claude Desktop**

3. **Start Using!**

Try these queries:
- "What CFPs are closing in the next 7 days?"
- "Show me Java-related conferences with open CFPs"
- "Find CFPs for conferences in France"
- "List AI conference CFPs closing this week"

### Mode 2: HTTP Server - For Cloud Deployment

Deploy as a web service that anyone can connect to without installing anything!

#### Deploy to Google Cloud Run

1. **Build and push the container:**
```bash
# Set your GCP project
gcloud config set project YOUR_PROJECT_ID

# Build and deploy in one command
gcloud run deploy developer-events-mcp \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --memory 256Mi
```

2. **Get your service URL:**
```bash
gcloud run services describe developer-events-mcp --region us-central1 --format 'value(status.url)'
```

3. **Connect from Claude Desktop:**

Update your config to use the HTTP endpoint:

```json
{
  "mcpServers": {
    "developer-events": {
      "url": "https://your-service-url.run.app"
    }
  }
}
```

#### Local HTTP Mode

Run locally as HTTP server for testing:
```bash
MODE=http PORT=8080 ./mcp-server
# Or let Cloud Run set PORT automatically
PORT=8080 ./mcp-server
```

#### Docker Build

```bash
docker build -t developer-events-mcp .
docker run -p 8080:8080 developer-events-mcp
```

#### Other Cloud Platforms

The Docker image works on any platform that supports containers:
- **AWS ECS/Fargate**: Deploy the container
- **Azure Container Apps**: `az containerapp create`
- **Fly.io**: `fly launch`
- **Railway**: Connect GitHub repo

## 🔧 Available Tools

### 1. `list_open_cfps`
List all currently open Call for Papers for developer conferences.

**Parameters:**
- `limit` (optional, default: 20): Maximum number of results

**Example:**
> Show me the next 10 open CFPs

### 2. `search_cfps_by_keyword`
Search open CFPs by keywords found in conference names and CFP links. Good for finding conferences related to technologies (e.g., 'java', 'python', 'kubernetes') or conference series names.

**Parameters:**
- `keywords` (required): List of keywords to search for in conference names and links
- `limit` (optional, default: 20): Maximum number of results

**Examples:**
> Find open CFPs with "Java" or "JVM" in the name

> Show me conferences with "AI" or "machine learning" in the title

> Search for CFPs related to Kubernetes

### 3. `find_closing_cfps`
Find CFPs that are closing soon within a specified number of days, optionally filtered by keywords.

**Parameters:**
- `daysAhead` (optional, default: 7): Number of days to look ahead
- `keywords` (optional): Filter by keywords in conference names

**Examples:**
> What CFPs are closing in the next 7 days?

> Show me Java-related CFPs closing in the next 14 days

> Find AI conference CFPs closing this week

### 4. `search_cfps_by_location`
Search for open CFPs by location/country.

**Parameters:**
- `location` (required): Location to search for (e.g., 'France', 'USA', 'Berlin', 'Online')
- `limit` (optional, default: 20): Maximum number of results

**Examples:**
> Find CFPs for conferences in France

> Show me online conferences with open CFPs

> What conferences in Berlin have open CFPs?

## 🛠️ Building from Source

### Prerequisites
- Go 1.21 or later

### Build for Current Platform
```bash
go build -ldflags="-s -w" -o developer-events-mcp main.go
```

### Build for All Platforms
```bash
./build-all.sh
```

This creates binaries in `bin/` for:
- Linux (x86_64, ARM64)
- macOS (Intel, Apple Silicon)
- Windows (x86_64)

## 📊 Performance

| Metric | Value |
|--------|-------|
| Binary Size | ~7MB |
| Startup Time | <10ms |
| Memory Usage | ~10-20MB |
| Runtime Dependencies | **NONE** |

## 🏗️ Architecture

- **Language:** Go 1.24
- **MCP SDK:** github.com/modelcontextprotocol/go-sdk v1.2.0
- **Transport:** Dual-mode
  - stdio (standard input/output) for local Claude Desktop
  - HTTP streaming (streamable HTTP) for cloud deployment
- **Caching:** In-memory (1 hour TTL)
- **HTTP Client:** Native Go net/http
- **Container:** Multi-stage Docker build (~15MB final image)

## 📡 Data Source

- API: https://developers.events/all-cfps.json
- Community-driven, updated regularly
- Currently ~2,785 total CFPs, ~305 open
- Cached for 1 hour to reduce API load

## 🔒 Security

- No external dependencies beyond Go standard library and MCP SDK
- HTTPS-only API calls
- No data stored persistently
- Runs in user space (no elevated privileges needed)
- Statically compiled binary

## 🐛 Troubleshooting

### Binary won't execute (macOS)
macOS may quarantine downloaded binaries. Remove the quarantine attribute:
```bash
xattr -d com.apple.quarantine mcp-server
```

### Permission denied (Linux/macOS)
Make the binary executable:
```bash
chmod +x mcp-server
```

### Claude Desktop doesn't see server
- Verify you're using the **full absolute path** in the config
- Restart Claude Desktop after config changes
- Check Claude Desktop logs:
  - macOS: `~/Library/Logs/Claude/`
  - Windows: `%APPDATA%\Claude\logs\`

### Server seems slow on first request
The first request fetches and caches data from developers.events API. Subsequent requests will be instant (served from 1-hour cache).

## 📝 Use Cases

### Daily Reminder Agent
Create a daily routine to check CFPs:
> Good morning! Check for Java, Kubernetes, or Cloud-related CFPs closing in the next 10 days

### Speaker Profile Matching
Set up topic filters that match your speaking profile:
> Find CFPs for: kubernetes, docker, cloud, devops, terraform

### Location-Based Planning
Plan conference speaking tours:
> Show me all open CFPs in Europe

> Find CFPs for conferences in USA happening in Q2 2026

### Last-Minute Opportunities
Find urgent opportunities:
> What CFPs are closing in the next 3 days?

## 🌍 Supported Platforms

Pre-built binaries available for:
- ✅ Linux x86_64
- ✅ Linux ARM64 (Raspberry Pi, AWS Graviton, etc.)
- ✅ macOS x86_64 (Intel)
- ✅ macOS ARM64 (Apple Silicon M1/M2/M3)
- ✅ Windows x86_64

## 📚 Response Format

Each CFP result includes:

```json
{
  "conference": "Conference Name",
  "location": "City (Country)",
  "cfpDeadline": "MMM-DD-YYYY",
  "daysRemaining": 30,
  "conferenceDate": "YYYY-MM-DD",
  "cfpLink": "https://...",
  "conferenceWebsite": "https://...",
  "status": "open"
}
```

## 🤝 Contributing

Contributions welcome! Please:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## 📄 License

MIT

## 🙏 Credits

- **MCP Protocol:** Anthropic
- **Go SDK:** Model Context Protocol team & Google
- **API Data:** developers.events community
- **Built with:** Go, official MCP SDK

---

**Built with ❤️ using Go and the official Model Context Protocol SDK**

Need help? Open an issue or check the troubleshooting section above.
