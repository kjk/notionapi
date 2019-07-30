#!/bin/bash
set -u -e -o pipefail -o verbose

go build -o tohtml github.com/kjk/notionapi/cmd/tohtml
rm -rf ./tohtml

go build -o dl github.com/kjk/notionapi/cmd/dl
rm -rf ./dl

go test -c github.com/kjk/notionapi
