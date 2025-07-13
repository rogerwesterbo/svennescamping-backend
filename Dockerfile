# Build the application binary
FROM --platform=$BUILDPLATFORM docker.io/golang:1.24.5-alpine AS builder

# Install ca-certificates and git for secure HTTPS and module downloads
RUN apk add --no-cache ca-certificates git

# Build arguments for multi-arch support
ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /workspace

# Copy dependency files first for better layer caching
COPY go.mod go.sum ./

# Download dependencies with security checks
RUN go mod download && go mod verify

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/

# Build the binary with security hardening flags
RUN CGO_ENABLED=0 \
    GOOS=$TARGETOS \
    GOARCH=$TARGETARCH \
    go build \
    -a \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -o svennescamping-backend \
    ./cmd/api/main.go

# Verify the binary was built successfully
RUN chmod +x svennescamping-backend

# Use distroless nonroot image for minimal attack surface
FROM gcr.io/distroless/static:nonroot

# Add labels for better image metadata
LABEL org.opencontainers.image.title="Svennes Camping Backend"
LABEL org.opencontainers.image.description="Backend service for Svennes Camping reservation system"
LABEL org.opencontainers.image.vendor="rogerwesterbo"
LABEL org.opencontainers.image.source="https://github.com/rogerwesterbo/svennescamping-backend"

# Create app directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /workspace/svennescamping-backend ./

# Copy CA certificates for HTTPS requests
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Use nonroot user (UID 65532)
USER 65532:65532

# Expose default port (can be overridden)
EXPOSE 8080

ENTRYPOINT ["/app/svennescamping-backend"]