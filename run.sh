#!/bin/bash
set -u -e -o pipefail -o verbose

go build github.com/kjk/notion/cmd/tohtml
./tohtml || true
rm -rf ./tohtml
