FROM alpine:3.18.4

WORKDIR /app
RUN apk add --update --no-cache postgresql-client curl && \
    rm -rf /var/cache/apk/*

COPY go-cron /usr/local/bin/go-cron
#COPY mc /usr/local/bin/mc

RUN curl https://dl.min.io/client/mc/release/linux-amd64/mc \
    --create-dirs \
    -o /usr/local/bin/mc

RUN chmod +x  /usr/local/bin/mc && chmod +x /usr/local/bin/go-cron

COPY run.sh backup.sh ./
RUN chmod +x run.sh && chmod +x backup.sh

RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

CMD ["sh", "run.sh"]
