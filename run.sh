#!/bin/bash
set -u -e -o pipefail

mkdir -p log
go build github.com/kjk/notionapi/cmd/tohtml
# https://www.notion.so/kjkpublic/Test-page-c969c9455d7c4dd79c7f860f3ace6429
# https://www.notion.so/kjkpublic/Test-page-text-4c6a54c68b3e4ea2af9cfaabcc88d58d
# https://www.notion.so/kjkpublic/Test-page-text-not-simple-f97ffca91f8949b48004999df34ab1f7
# https://www.notion.so/kjkpublic/blog-300db9dc27c84958a08b8d0c37f4cfe5

# c969c9455d7c4dd79c7f860f3ace6429   test all
# f97ffca91f8949b48004999df34ab1f7   test text not simple
# 6682351e44bb4f9ca0e149b703265bdb   test header
# fd9338a719a24f02993fcfbcf3d00bb0   test todo list and page style
# 484919a1647144c29234447ce408ff6b   test toggle and bullet list
# c969c9455d7c4dd79c7f860f3ace6429
# 300db9dc27c84958a08b8d0c37f4cfe5   large page (my blog)
# 0367c2db381a4f8b9ce360f388a6b2e3   index page for test pages
# 25b6ac21d67744f18a4dc071b21a86fe   test code and favorite
# 70ecbf1f5abc41d48a4e4320aeb38d10   test todo
# 97100f9c17324fd7ba3d3c5f1832104d   test dates
# 0fa8d15a16134f0c9fad1aa0a7232374   test comments, icon, cover
# 57cb49183ee44eb9a4fcc37817473b54   test deleted page
# 157765353f2c4705bd45474e5ba8b46c   notion "what's new" page
# 72fd504c58984cc5a5dfb86b6f8617dc   test nested toggle

# available args:
# -recursive -no-cache
./tohtml -no-cache fa3fc358e5644f39b89c57f13d426d54 || true
rm -rf ./tohtml
