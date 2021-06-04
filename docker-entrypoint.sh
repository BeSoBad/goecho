#!/bin/bash
set -e

case "$1" in
    start)
        /app/echo "$2"
        ;;
    *)
        exec "$@"
esac