#!/bin/bash
set -e

case "$1" in
    start)
        exec /app/echo
        ;;
    *)
        exec "$@"
esac