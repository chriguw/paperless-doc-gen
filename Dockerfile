# ── Build stage ──────────────────────────────────────────────────────────────
FROM --platform=$BUILDPLATFORM golang:1.26-alpine AS builder

WORKDIR /app

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary for target platform
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o dochandler cmd/dochandler/main.go

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM --platform=$TARGETPLATFORM alpine:3.19

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/dochandler .

# Copy fonts
COPY fonts/ ./fonts/

# Create reports directory
RUN mkdir -p reports

EXPOSE 8080

CMD ["./dochandler"]

