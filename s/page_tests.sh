#!/bin/bash
set -u -e -o pipefail

go build github.com/kjk/notionapi/cmd/tests
./tests || true
rm -rf ./tests
