module github.com/kjk/notionapi/do

require (
	github.com/kjk/fmthtml v0.0.0-20190816041536-39f5e479d32d
	github.com/kjk/notionapi v0.0.0-20210703151502-e954b9b9578e
	github.com/kjk/u v0.0.0-20210327060556-13ea33918991
	github.com/klauspost/cpuid/v2 v2.0.9 // indirect
	github.com/kr/pretty v0.1.0 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97 // indirect
	golang.org/x/net v0.0.0-20210805182204-aaa1db679c0d // indirect
	golang.org/x/sys v0.0.0-20210630005230-0f9fa26af87c // indirect
	gopkg.in/check.v1 v1.0.0-20180628173108-788fd7840127 // indirect
)

replace github.com/kjk/notionapi => ./..

go 1.13
