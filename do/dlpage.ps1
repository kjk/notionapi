
#!/bin/env pwsh

# Usage example:
# .\s\dlpage.ps1 4c6a54c68b3e4ea2af9cfaabcc88d58d
#
# Options:
#   -use-cache : use on-disk cache to maybe avoid downloading
#                data from the server

# For testing: downloads a page with a given notion id
# and saves http requests and responses in
# log/${notionid}.log.txt so that we can look at them
# It will also save log/${notionid}.page.json which is
# JSON-serialized Page structure.

Remove-Item -Force -ErrorAction Ignore ./test.exe
go build -o test.exe github.com/kjk/notionapi/cmd/test
./test.exe -dlpage $args
Remove-Item -Force -ErrorAction Ignore ./test.exe
