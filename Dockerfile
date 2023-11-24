FROM golang:1.21-alpine as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o pg2minio ./main.go

RUN go install github.com/minio/mc@latest

FROM alpine:3.18

RUN apk add --update --no-cache postgresql-client && \
    rm -rf /var/cache/apk/*

WORKDIR /app

COPY --from=builder /app/pg2minio /usr/local/bin/pg2minio
RUN chmod +x /usr/local/bin/pg2minio

COPY --from=builder /go/bin/mc /usr/local/bin/mc
RUN chmod +x /usr/local/bin/mc

RUN chmod 0777 /app

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD ["/usr/local/bin/pg2minio"]
#ENTRYPOINT ["tail", "-f", "/dev/null"]
