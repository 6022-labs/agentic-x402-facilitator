#!/usr/bin/env sh
set -e

if [ -n "${PORT}" ] && [ -z "${GF_SERVER_HTTP_PORT}" ]; then
  export GF_SERVER_HTTP_PORT="${PORT}"
fi

exec /run.sh
