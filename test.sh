#!/bin/bash
set -e

rm -f *.dat
NSQD_LOGFILE=$(mktemp -t nsqlookupd.XXXXXXX)
echo "  logging to $NSQD_LOGFILE"

export HOST_NAME=void

docker-compose --file ./docker-compose.yaml up >$NSQD_LOGFILE 2>&1 &

NSQD_PID=$!
echo "started nsqd PID $NSQD_PID"

sleep 1

go test -v -timeout 60s ./...

cleanup() {
    echo "killing nsqd PID $NSQD_PID"
    kill -s TERM $NSQD_PID || cat $NSQD_LOGFILE
}
trap cleanup INT TERM EXIT
