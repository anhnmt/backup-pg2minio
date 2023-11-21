FROM golang:1.21-alpine as builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o go-cron ./cmd/go-cron/main.go

# Or use go install
# RUN go install github.com/anhnmt/backup-pg2minio/cmd/go-cron@latest
RUN go install github.com/minio/mc@latest

FROM alpine:3.18

WORKDIR /app
RUN apk add --update --no-cache postgresql-client curl bash && \
    rm -rf /var/cache/apk/*

COPY --from=builder /app/go-cron /usr/local/bin/go-cron
COPY --from=builder /go/bin/mc /usr/local/bin/mc
RUN chmod +x /usr/local/bin/mc && chmod +x /usr/local/bin/go-cron

COPY run.sh backup.sh ./
RUN chmod +x run.sh && chmod +x backup.sh

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD ["bash", "run.sh"]
# ENTRYPOINT ["tail", "-f", "/dev/null"]
