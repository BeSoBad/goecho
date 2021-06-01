#!/bin/bash
set -e

case "$1" in
    start)
        /app/goecho "$2"
        ;;
    *)
        exec "$@"
esac