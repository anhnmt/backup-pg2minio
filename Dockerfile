ARG GO_VERSION=${GO_VERSION:-"1.25"}
ARG ALPINE_VERSION=${ALPINE_VERSION:-"3.21"}

FROM postgres:18-alpine${ALPINE_VERSION} AS base

RUN apk add --update --no-cache \
			ca-certificates \
			curl \
		&& curl -Lo /usr/bin/mc https://dl.min.io/client/mc/release/linux-amd64/mc \
		&& chmod +x /usr/bin/mc \
    && rm -rf /var/log/* \
    && rm -rf /var/cache/apk/*

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION} AS builder
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pg2minio ./main.go

FROM base

WORKDIR /app

COPY --from=builder /app/pg2minio /usr/local/bin/pg2minio
RUN chmod +x /usr/local/bin/pg2minio

RUN chmod 0777 /app
RUN chmod 0777 /usr/local/bin

CMD ["/usr/local/bin/pg2minio"]
