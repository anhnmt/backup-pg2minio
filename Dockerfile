ARG GO_VERSION=${GO_VERSION:-"1.24"}
ARG ALPINE_VERSION=${ALPINE_VERSION:-"3.21"}

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pg2minio ./main.go

FROM alpine:${ALPINE_VERSION} AS libs
ARG POSTGRES_CLIENT_VERSION
ARG MINIO_CLIENT_VERSION

RUN apk add --update --no-cache \
    	postgresql17-client \
			minio-client \
			ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/log/* \
    && rm -rf /var/cache/apk/*

FROM scratch
WORKDIR /app

COPY --from=libs / /
COPY --from=builder /app/pg2minio /usr/local/bin/pg2minio
RUN chmod +x /usr/local/bin/pg2minio

RUN chmod 0777 /app
RUN chmod 0777 /usr/local/bin

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD ["/usr/local/bin/pg2minio"]
# ENTRYPOINT ["tail", "-f", "/dev/null"]
