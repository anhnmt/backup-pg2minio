ARG GO_VERSION=1.26
ARG ALPINE_VERSION=3.23
ARG PG_VERSION=18

# ── Stage 1: Download mc binary ───────────────────────────────────────────────
FROM alpine:${ALPINE_VERSION} AS mc-downloader

RUN apk add --no-cache curl \
    && curl -Lo /usr/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc \
    && chmod +x /usr/bin/mc

# ── Stage 2: Build Go binary ──────────────────────────────────────────────────
FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-s -w -extldflags '-static'" \
    -trimpath \
    -o pg2minio \
    ./main.go

# ── Stage 3: Compress binary ──────────────────────────────────────────────────
FROM alpine:${ALPINE_VERSION} AS compressor

RUN apk add --no-cache upx

COPY --from=builder /app/pg2minio /pg2minio

RUN upx --best --lzma /pg2minio \
    && upx --test /pg2minio

# ── Stage 4: Final ────────────────────────────────────────────────────────────
FROM alpine:${ALPINE_VERSION}

ARG PG_VERSION

# Only install pg_dump client, not the full PostgreSQL server
RUN apk add --no-cache \
        ca-certificates \
        postgresql${PG_VERSION}-client \
    && rm -rf /var/cache/apk/*

COPY --from=mc-downloader /usr/bin/mc /usr/bin/mc
COPY --from=compressor /pg2minio /usr/local/bin/pg2minio

WORKDIR /app

CMD ["/usr/local/bin/pg2minio"]