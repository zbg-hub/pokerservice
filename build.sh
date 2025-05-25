#!/bin/bash

RUN_NAME="pokerservice"

mkdir -p output/bin output/conf
cp script/bootstrap.sh script/settings.py output
chmod +x output/bootstrap.sh
cp -r conf/* output/conf 2>/dev/null

export GO111MODULE="on"
go build -ldflags -v -o output/bin/${RUN_NAME}
