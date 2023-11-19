#!/bin/sh

set -e

if [[ -z "${SCHEDULE}" ]]; then
  sh backup.sh
else
  exec go-cron "$SCHEDULE" /bin/sh backup.sh
fi