#!/bin/bash
set -u -e -o pipefail

mkdir -p log
go build github.com/kjk/notion/cmd/tohtml
# https://www.notion.so/kjkpublic/Test-page-c969c9455d7c4dd79c7f860f3ace6429
# https://www.notion.so/kjkpublic/Test-page-text-4c6a54c68b3e4ea2af9cfaabcc88d58d
# https://www.notion.so/kjkpublic/Test-page-text-not-simple-f97ffca91f8949b48004999df34ab1f7
# https://www.notion.so/kjkpublic/blog-300db9dc27c84958a08b8d0c37f4cfe5

# f97ffca91f8949b48004999df34ab1f7   test text not simple
# 6682351e44bb4f9ca0e149b703265bdb   test header
# fd9338a719a24f02993fcfbcf3d00bb0   test todo list
# 484919a1647144c29234447ce408ff6b   test toggle and bullet list
# c969c9455d7c4dd79c7f860f3ace6429
# 300db9dc27c84958a08b8d0c37f4cfe5   large page (my blog)
# 0367c2db381a4f8b9ce360f388a6b2e3   index page for test pages
# 25b6ac21d67744f18a4dc071b21a86fe   test code
# 70ecbf1f5abc41d48a4e4320aeb38d10   test todo

# -recursive -no-cache
./tohtml -no-cache 70ecbf1f5abc41d48a4e4320aeb38d10 || true
rm -rf ./tohtml
