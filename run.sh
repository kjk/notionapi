#!/bin/bash
set -u -e -o pipefail -o verbose

mkdir -p log
go build github.com/kjk/notion/cmd/tohtml
./tohtml || true
rm -rf ./tohtml
