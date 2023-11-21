#!/bin/bash

set -e

if [[ -z "${SCHEDULE}" ]]; then
  bash backup.sh
else
  exec go-cron "$SCHEDULE" /bin/bash backup.sh
fi