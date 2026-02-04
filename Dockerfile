# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY main.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o developer-events-mcp main.go

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/developer-events-mcp .

# Run as non-root user for security
RUN adduser -D -u 1000 mcpuser
USER mcpuser

# Cloud Run sets PORT environment variable
EXPOSE 8080

# Run the server in HTTP mode
CMD ["./developer-events-mcp"]
