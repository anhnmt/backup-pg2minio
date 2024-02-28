FROM golang:1.22-alpine as builder

ENV MINIO_CLIENT_VERSION=RELEASE.2024-02-24T01-33-20Z

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pg2minio ./main.go

RUN go install github.com/minio/mc@${MINIO_CLIENT_VERSION}

FROM alpine:3.19 as libs

ENV POSTGRES_CLIENT_VERSION=16.2-r0

RUN apk update && apk upgrade

RUN apk add --update --no-cache postgresql16-client=${POSTGRES_CLIENT_VERSION} ca-certificates \
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/log/* \
    && rm -rf /var/cache/apk/*

FROM scratch

WORKDIR /app

COPY --from=libs / /

COPY --from=builder /app/pg2minio /usr/local/bin/pg2minio
RUN chmod +x /usr/local/bin/pg2minio

COPY --from=builder /go/bin/mc /usr/local/bin/mc
RUN chmod +x /usr/local/bin/mc

RUN chmod 0777 /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD ["/usr/local/bin/pg2minio"]
#ENTRYPOINT ["tail", "-f", "/dev/null"]
