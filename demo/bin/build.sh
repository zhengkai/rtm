#!/bin/bash -e

cd $(dirname `readlink -f $0`)

BIN="server-${1:-dev}"

TIME="time: %E" time go build -o "$BIN" ../*.go

"./${BIN}"
