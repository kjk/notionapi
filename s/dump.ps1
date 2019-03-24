#!/bin/env pwsh

Remove-Item -Force -ErrorAction Ignore ./dump.exe
go build -o dump.exe github.com/kjk/notionapi/cmd/dump
./dump.exe $args
Remove-Item -Force -ErrorAction Ignore ./dump.exe
