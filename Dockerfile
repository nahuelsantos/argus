# Build stage
FROM golang:1.21-alpine AS builder

# Build arguments (optional)
ARG VERSION=dev
ARG BUILD_TIME=unknown
ARG GIT_COMMIT=unknown

# Set working directory
WORKDIR /app

# Install git for potential private dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X 'github.com/nahuelsantos/argus/internal/config.Version=${VERSION}' \
              -X 'github.com/nahuelsantos/argus/internal/config.BuildTime=${BUILD_TIME}' \
              -X 'github.com/nahuelsantos/argus/internal/config.GitCommit=${GIT_COMMIT}'" \
    -a -installsuffix cgo -o argus ./cmd/argus/

# Final stage
FROM alpine:3.19

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create non-root user
RUN addgroup -g 1001 -S argus && \
    adduser -u 1001 -S argus -G argus

WORKDIR /root/

# Copy the binary and static files
COPY --from=builder /app/argus .
COPY --from=builder /app/static ./static

# Change ownership to non-root user
RUN chown -R argus:argus /root

# Switch to non-root user
USER argus

# Expose port
EXPOSE 3001

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:3001/health || exit 1

# Set labels for GHCR
LABEL org.opencontainers.image.title="Argus"
LABEL org.opencontainers.image.description="LGTM Stack Validator - The All-Seeing LGTM Stack Testing & Validation Tool"
LABEL org.opencontainers.image.vendor="nahuelsantos"
LABEL org.opencontainers.image.licenses="MIT"
LABEL org.opencontainers.image.url="https://github.com/nahuelsantos/argus"
LABEL org.opencontainers.image.source="https://github.com/nahuelsantos/argus"
LABEL org.opencontainers.image.documentation="https://github.com/nahuelsantos/argus/blob/main/README.md"
LABEL org.opencontainers.image.version="v0.0.1"

# Run the binary
CMD ["./argus"] 