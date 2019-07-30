
#!/bin/env pwsh

# Usage example:
# .\s\tohtml.ps1 4c6a54c68b3e4ea2af9cfaabcc88d58d
#
# Options:
#   -no-open   : won't automatically open the web browser
#   -use-cache : use on-disk cache to maybe avoid downloading
#                data from the server

# For testing: downloads a page with a given notion id
# and converts to HTML. Saves the html to log/ directory
# and opens browser with that page

Remove-Item -Force -ErrorAction Ignore ./test.exe
go build -o test.exe github.com/kjk/notionapi/cmd/test
./test.exe -tohtml $args
Remove-Item -Force -ErrorAction Ignore ./test.exe
