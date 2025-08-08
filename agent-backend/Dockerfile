# Use the official Go image as the base image (pinned by digest for reproducibility)
# To get the actual digest: docker pull golang:1.24.2-alpine && docker images --digests | grep golang
FROM golang:1.24.2-alpine@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef AS builder

# Build arguments for reproducible builds
ARG VERSION_SHA
ARG BUILD_DATE
ARG VCS_REF

# Set the working directory
WORKDIR /app

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application with deterministic flags
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -buildvcs=false \
    -ldflags="-s -w" \
    -a -installsuffix cgo \
    -o agent ./cmd/agent

# Use a minimal alpine image for the final stage (pinned by digest for reproducibility)
# To get the actual digest: docker pull alpine:3.19 && docker images --digests | grep alpine
FROM alpine:3.19@sha256:1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef

# Build arguments for reproducible builds
ARG VERSION_SHA
ARG BUILD_DATE
ARG VCS_REF

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates curl

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/agent .

# Copy the config file
COPY --from=builder /app/config.yaml .

# Change ownership to the non-root user
RUN chown -R appuser:appgroup /app

# Add OCI labels for reproducible builds
LABEL org.opencontainers.image.title="Pandacea Agent Backend" \
      org.opencontainers.image.description="P2P node and HTTP API server for the Pandacea Protocol" \
      org.opencontainers.image.vendor="Pandacea Protocol" \
      org.opencontainers.image.version="${VERSION_SHA:-latest}" \
      org.opencontainers.image.created="${BUILD_DATE}" \
      org.opencontainers.image.revision="${VCS_REF}" \
      org.opencontainers.image.source="https://github.com/pandacea/pandacea-protocol" \
      org.opencontainers.image.licenses="MIT"

# Switch to the non-root user
USER appuser

# Expose the port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=40s --retries=3 \
  CMD curl -f http://localhost:8080/healthz || exit 1

# Run the application
CMD ["./agent"] 