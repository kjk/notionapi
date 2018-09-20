#!/bin/bash
set -u -e -o pipefail -o verbose

go run cmd/dl/*.go $@
