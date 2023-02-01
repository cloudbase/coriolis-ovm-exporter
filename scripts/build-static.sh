#!/bin/sh

EXPORTER_SOURCE="/build/coriolis-ovm-exporter"

cd $EXPORTER_SOURCE/cmd/exporter
go build -buildvcs=false -o $EXPORTER_SOURCE/coriolis-ovm-exporter -ldflags "-linkmode external -extldflags '-static' -s -w -X main.Version=$(git describe --always --dirty)" .
