## 🐳 Docker

### Prerequisites

- [Docker Desktop](https://www.docker.com/products/docker-desktop/) installed
- [Docker Hub](https://hub.docker.com/) account

---

### Setup Buildx (only needed once)

```bash
# Create and use a new buildx builder
docker buildx create --name multibuilder --use

# Bootstrap and verify
docker buildx inspect --bootstrap
```

You should see `linux/amd64` and `linux/arm64` in the platforms list. If `linux/arm64` is missing install QEMU:

---

### Build and Push to Docker Hub

```bash
# Login to Docker Hub
docker login

# Build for both platforms and push
docker buildx build --platform linux/amd64,linux/arm64 -t yourusername/dochandler:latest -t yourusername/dochandler:1.0.0 --push .
```

> ⚠️ Always run from the **project root** where the `Dockerfile` is located.
> ⚠️ The `.` at the end is required — it defines the build context.

---

### Build Locally for Testing (single platform)

```bash
# amd64
docker buildx build --platform linux/amd64 -t dochandler:latest --load .

# arm64
docker buildx build --platform linux/arm64 -t dochandler:latest --load .

# Verify
docker images | grep dochandler
```

> ℹ️ `--load` only supports a single platform at a time and stores the image locally.
> ℹ️ `--push` supports multiple platforms but sends directly to Docker Hub — not visible in `docker images`.

---

### Verify Published Image

```bash
docker buildx imagetools inspect yourusername/dochandler:latest
```

Expected output:
```
Name:      docker.io/yourusername/dochandler:latest
MediaType: application/vnd.oci.image.index.v1+json

Manifests:
  Platform:  linux/amd64    ✅
  Platform:  linux/arm64    ✅
```

---

### Tagging Strategy

| Tag | Meaning |
|---|---|
| `latest` | Always points to the most recent build |
| `1.0.0` | Exact version — never changes |
| `1.0` | Latest patch of version 1.0 |
| `1` | Latest minor/patch of major version 1 |

```bash
docker buildx build --platform linux/amd64,linux/arm64 -t yourusername/dochandler:latest -t yourusername/dochandler:1.0.0 -t yourusername/dochandler:1.0 -t yourusername/dochandler:1 --push .
```

---

### Platform Detection

With multi-platform images you **do not need separate tags** for ARM and AMD. Docker automatically pulls the correct platform:

| Machine | Platform pulled |
|---|---|
| Intel / AMD server | `linux/amd64` |
| Raspberry Pi 4 / 5 | `linux/arm64` |
| Apple Silicon Mac | `linux/arm64` |
| AWS Graviton | `linux/arm64` |

---

### docker-compose.yml (standalone)

```yaml
version: "3.8"

services:
  dochandler:
    image: yourusername/dochandler:latest
    container_name: dochandler
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - PAPERLESS_URL=http://paperless-ngx:8000
      - PAPERLESS_TOKEN=your-api-token-here
    volumes:
      - ./reports:/app/reports
```

---

### Integration into Paperless docker-compose.yml

Add the `dochandler` service to your existing Paperless `docker-compose.yml`:

```yaml
services:
 broker:
   image: redis:8
   container_name: broker
   restart: unless-stopped
   volumes:
     - /Users/yourusername/paperless/data/paperless/redis/_data:/data

 db:
   image: postgres:18
   container_name: db
   restart: unless-stopped
   volumes:
     - /Users/yourusername/paperless/data/paperless/postgresql/_data:/var/lib/postgresql
   environment:
     POSTGRES_DB: paperless
     POSTGRES_USER: paperless
     POSTGRES_PASSWORD: paperless

 webserver:
   image: ghcr.io/paperless-ngx/paperless-ngx:2.20.15
   container_name: webserver
   restart: unless-stopped
   depends_on:
     - db
     - broker
     - gotenberg
     - tika
   ports:
     - "8001:8000"
   volumes:
     - /Users/yourusername/paperless/data/paperless/consume:/usr/src/paperless/consume
     - /Users/yourusername/paperless/data/paperless/data:/usr/src/paperless/data
     - /Users/yourusername/paperless/data/paperless/media:/usr/src/paperless/media
     - /Users/yourusername/paperless/data/paperless/export:/usr/src/paperless/export
   environment:
     PAPERLESS_REDIS: redis://broker:6379
     PAPERLESS_DBHOST: db
     PAPERLESS_TIKA_ENABLED: 1
     PAPERLESS_TIKA_GOTENBERG_ENDPOINT: http://gotenberg:3000
     PAPERLESS_TIKA_ENDPOINT: http://tika:9998
     PAPERLESS_OCR_LANGUAGE: deu
     PAPERLESS_TIME_ZONE: Europe/Berlin
     PAPERLESS_CONSUMER_ENABLE_BARCODES: "true"
     PAPERLESS_CONSUMER_ENABLE_ASN_BARCODE: "true"
     PAPERLESS_CONSUMER_BARCODE_SCANNER: ZXING
     PAPERLESS_CONSUMER_POLLING: 60
     PAPERLESS_EMAIL_TASK_CRON: '*/10 * * * *'
     PAPERLESS_WORKFLOW_ENABLED: "true"
     PAPERLESS_URL: http://localhost:8081
     USERMAP_UID: "1002"
     USERMAP_GID: "1002"

 gotenberg:
   image: gotenberg/gotenberg:8.8
   restart: unless-stopped
   command:
     - "gotenberg"
     - "--chromium-disable-javascript=false"
     - "--chromium-allow-list=.*"

 tika:
   image: docker.io/apache/tika:latest
   container_name: tika
   restart: unless-stopped

 dochandler:
    image: yourusername/dochandler:latest
    container_name: dochandler
    restart: unless-stopped
    ports:
      - "8082:8080"
    environment:
      - PAPERLESS_URL=http://webserver:8000   # use the internal docker service name
      - PAPERLESS_TOKEN=your-paperless-token
    volumes:
      - /Users/yourusername/paperless/data/paperless/consume:/app/reports     # reports are accessible on the host
    depends_on:
      - webserver
```

> ℹ️ `PAPERLESS_URL` must use the **internal Docker service name** (`webserver`) not `localhost`.
> ℹ️ The reports volume is mounted to the Paperless `consume` folder so generated PDFs are automatically imported.

---

### Dockerfile

```dockerfile
# ── Build stage ───────────────────────────────────────────────────────────────
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
FROM alpine:3.19

WORKDIR /app

COPY --from=builder /app/dochandler .
COPY fonts/ ./fonts/
RUN mkdir -p reports

EXPOSE 8080

CMD ["./dochandler"]
```

---

### .dockerignore

```
# Git
.git
.gitignore

# Reports
reports/

# Binaries
bin/
*.exe

# IDE
.vscode/
.idea/

# Env
.env
```
