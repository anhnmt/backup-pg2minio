# backup-pg2minio
Backup Postgresql to Minio

# Supported Versions
* 11
* 12
* 13
* 14
* 15
* 16
* 17

#### `cp .env.example .env`
```
MINIO_ACCESS_KEY=12345678
MINIO_SECRET_KEY=12345678
MINIO_BUCKET=minio
MINIO_SERVER=http://127.0.0.1:9000
```

#### `docker-compose up -d`
```
version: '3'
services:
  pg2minio:
    container_name: pg2minio
    image: anhnmt/backup-pg2minio:latest
    build:
      context: .
      dockerfile: ./Dockerfile
    env_file:
      - .env
    network_mode: host
```

## Required Environment Variables

- `MINIO_ACCESS_KEY` - Your Minio access key.
- `MINIO_SECRET_KEY` - Your Minio access key.
- `MINIO_BUCKET` - Your Minio bucket.
- `MINIO_SERVER` - Your Minio server

- `POSTGRES_HOST` - Hostname of the PostgreSQL database to backup, alternatively this container can be linked to the container with the name `postgres`.

### Optional Environment Variables

- `SCHEDULE` - Cron schedule to run periodic backups.

- `POSTGRES_PASSWORD` - Password for the PostgreSQL user, if you are using a database on the same machine this isn't usually needed.
- `POSTGRES_PORT` - Port of the PostgreSQL database, uses the default `5432`.
- `POSTGRES_EXTRA_OPTS` - Extra arguments to pass to the `pg_dump` command.
- `POSTGRES_DATABASE` - Name of the PostgreSQL database to backup.
- `POSTGRES_USER` - PostgreSQL user, with priviledges to dump the database.
- `POSTGRES_PRERUN` - Check connection before executing.

- `MINIO_API_VERSION` - you can change with `S3v4` or `S3v2`.
- `MINIO_CLEAN` - Assign a value to activate, default is `0`. For example: `7d`, `14d`, `1m`, `30s`
- `MINIO_BACKUP_DIR` - Allows you to change the path in the bucket. e.g. abc/def (without / at the beginning and end)
- `MINIO_INSECURE` - Disables TLS/SSL certificate verification.
- `MINIO_DEBUG` - Enables verbose output to the console.
- `MINIO_PRERUN` - Check connection before executing.

- `TELEGRAM_ENABLED` - Set `true` to enable. Default is `false`
- `TELEGRAM_CHAT_ID` - Chat ID for example `-40054422`
- `TELEGRAM_TOKEN` - Telegram bot token for example `610864305:AAGw2BVSPYPjcc8940bswQTRUZIssSFJA`, without `bot` at the beginning

- `METRICS_ENABLED` - Set `true` to enable metrics collection. Default is `false`.
- `METRICS_NAMESPACE` - Namespace for the metrics. Default is empty.
- `METRICS_SUBSYSTEM` - Subsystem for the metrics. Default is empty.
- `METRICS_PORT` - Port for the metrics endpoint. Default is `8080`.
- `METRICS_PATH` - Path for the metrics endpoint. Default is `/metrics`.

# some script from 
-  alpine packages : https://pkgs.alpinelinux.org/packages
-  go-cron : https://github.com/michaloo/go-cron
-  docker-backup-postgres-s3 : https://github.com/wonderu/docker-backup-postgres-s3
-  postgres-backup-s3 : https://github.com/schickling/dockerfiles/tree/master/postgres-backup-s3 
-  MinIO Client : https://github.com/minio/mc
-  Cron parser : [https://elmah.io/tools/cron-parser](https://elmah.io/tools/cron-parser/#0_*/5_*_*_*_*)
-  More information about the scheduling can be found [here](http://godoc.org/github.com/robfig/cron#hdr-Predefined_schedules).
