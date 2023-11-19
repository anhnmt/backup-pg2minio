#!/bin/sh

set -e
set -o pipefail

if [[ -z "${MINIO_ACCESS_KEY}" ]]; then
  echo "You need to set the MINIO_ACCESS_KEY environment variable."
  exit 1
fi

if [[ -z "${MINIO_SECRET_KEY}" ]]; then
  echo "You need to set the MINIO_SECRET_KEY environment variable."
  exit 1
fi

if [[ -z "${MINIO_BUCKET}"  ]]; then
  echo "You need to set the MINIO_BUCKET environment variable."
  exit 1
fi

if [[ -z "${MINIO_SERVER}" ]]; then
  echo "You need to set the MINIO_SERVER environment variable."
  exit 1
fi

if [[ -z "${POSTGRES_DATABASE}" ]]; then
  echo "You need to set the POSTGRES_DATABASE environment variable."
  exit 1
fi

if [[ -z "${POSTGRES_HOST}"  ]]; then
  if [ -n "${POSTGRES_PORT_5432_TCP_ADDR}" ]; then
    POSTGRES_HOST=$POSTGRES_PORT_5432_TCP_ADDR
    POSTGRES_PORT=$POSTGRES_PORT_5432_TCP_PORT
  else
    echo "You need to set the POSTGRES_HOST environment variable."
    exit 1
  fi
fi

if [[ -z "${POSTGRES_USER}" ]]; then
  echo "You need to set the POSTGRES_USER environment variable."
  exit 1
fi

if [[ -z "${POSTGRES_PASSWORD}" ]]; then
  echo "You need to set the POSTGRES_PASSWORD environment variable or link to a container named POSTGRES."
  exit 1
fi

export MINIO_ACCESS_KEY=$MINIO_ACCESS_KEY
export MINIO_SECRET_KEY=$MINIO_SECRET_KEY
export MINIO_SERVER=$MINIO_SERVER
export MINIO_API_VERSION=$MINIO_API_VERSION

mc alias set minio "$MINIO_SERVER" "$MINIO_ACCESS_KEY" "$MINIO_SECRET_KEY" --api "$MINIO_API_VERSION" > /dev/null

export PGPASSWORD=$POSTGRES_PASSWORD
POSTGRES_HOST_OPTS="-h $POSTGRES_HOST -p $POSTGRES_PORT -U $POSTGRES_USER $POSTGRES_EXTRA_OPTS"

echo "Creating dump of ${POSTGRES_DATABASE} database from ${POSTGRES_HOST}..."

pg_dump $POSTGRES_HOST_OPTS $POSTGRES_DATABASE | gzip > $HOME/tmp_dump.sql.gz

TIMESTAMP=$(date +"%Y-%m-%dT%H:%M:%SZ")
MINIO_PATH="minio/$MINIO_BUCKET"
FILE_NAME="${POSTGRES_DATABASE}_${TIMESTAMP}.sql.gz"

function copy_to_minio {
  mc cp $1 $2 || exit 2
  rm $HOME/tmp_dump.sql.gz
  sync
}

MINIO_STORE_DIR="$MINIO_PATH/${POSTGRES_DATABASE}"

# check if have CUSTOM_DIR
if [[ -n "${CUSTOM_DIR}" ]]; then
  MINIO_STORE_DIR="$MINIO_PATH/${CUSTOM_DIR}/${POSTGRES_DATABASE}"
fi

echo "Minio store dir $MINIO_STORE_DIR"

echo "Uploading dump to $MINIO_BUCKET"

copy_to_minio "$HOME/tmp_dump.sql.gz" "$MINIO_STORE_DIR/$FILE_NAME"

echo "Upload successfully"

echo "Cleaning up..."

if [[ -n "${MINIO_CLEAN}" ]]; then
  if [ -z "$MINIO_CLEAN" ]; then
      MINIO_CLEAN=0
  fi

  mc find $MINIO_STORE_DIR --older-than $MINIO_CLEAN --exec "mc rm {}"
fi

echo "Cleanup successfully"

echo "SQL backup & cleanup successfully" 1>&2
