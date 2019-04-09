#!/bin/env pwsh

go test -c github.com/kjk/notionapi
go test -c github.com/kjk/notionapi/tohtml
go build -o tmp.exe github.com/kjk/notionapi/cmd/test
remove-item ./tmp.exe
