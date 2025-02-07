# Use multi-stage build for smaller final image
FROM golang:1.22-alpine3.19 AS builder

WORKDIR /app

# Install necessary build tools
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build for multiple architectures
RUN GOOS=linux GOARCH=amd64 go build -o public-api-scraper-amd64 .
RUN GOOS=linux GOARCH=arm64 go build -o public-api-scraper-arm64 .

# Final stage
FROM alpine:3.19

WORKDIR /app

# Copy binary based on architecture
COPY --from=builder /app/public-api-scraper-* ./
COPY --from=builder /app/templates ./templates

# Create script to select correct binary
RUN echo '#!/bin/sh' > /app/entrypoint.sh && \
    echo 'ARCH=$(uname -m)' >> /app/entrypoint.sh && \
    echo 'if [ "$ARCH" = "x86_64" ]; then' >> /app/entrypoint.sh && \
    echo '    exec ./public-api-scraper-amd64' >> /app/entrypoint.sh && \
    echo 'elif [ "$ARCH" = "aarch64" ]; then' >> /app/entrypoint.sh && \
    echo '    exec ./public-api-scraper-arm64' >> /app/entrypoint.sh && \
    echo 'else' >> /app/entrypoint.sh && \
    echo '    echo "Unsupported architecture: $ARCH"' >> /app/entrypoint.sh && \
    echo '    exit 1' >> /app/entrypoint.sh && \
    echo 'fi' >> /app/entrypoint.sh && \
    chmod +x /app/entrypoint.sh

# Set environment variables
ENV PORT=3000

# Expose port
EXPOSE 3000

# Run the application
ENTRYPOINT ["/app/entrypoint.sh"] 