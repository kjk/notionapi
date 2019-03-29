#!/bin/env pwsh

Remove-Item -Force -ErrorAction Ignore ./tests.exe

go build -o tests.exe github.com/kjk/notionapi/cmd/tests
./tests.exe

Remove-Item -Force -ErrorAction Ignore ./tests.exe
