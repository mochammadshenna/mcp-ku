# Build stage
FROM golang:1.24-alpine AS builder

# Set necessary environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

# Move to working directory /build
WORKDIR /build

# Copy and download dependency using go mod
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container
COPY . .

# Build the application
RUN go build -ldflags="-s -w" -o mcp-server ./cmd/server/main.go
RUN go build -ldflags="-s -w" -o mcp-client ./cmd/client/main.go

# Final stage
FROM alpine:latest

# Add ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

# Create appuser
RUN adduser -D -g '' appuser

WORKDIR /root/

# Copy the pre-built binary file and other necessary files
COPY --from=builder /build/mcp-server .
COPY --from=builder /build/mcp-client .
COPY --from=builder /build/migrations ./migrations/
COPY --from=builder /build/.env.example .env

# Change ownership to appuser
RUN chown -R appuser:appuser /root/

# Use an unprivileged user
USER appuser

# Expose port
EXPOSE 8080

# Command to run when starting the container
CMD ["./mcp-server"]