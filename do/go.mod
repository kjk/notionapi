module github.com/kjk/notionapi/do

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/kjk/atomicfile v0.0.0-20220410204726-989ae30d2b66 // indirect
	github.com/kjk/fmthtml v0.0.0-20190816041536-39f5e479d32d
	github.com/kjk/notionapi v0.0.0-20230925082132-8a7dbab354e9
	github.com/kjk/u v0.0.0-20220410204605-ce4a95db4475
	github.com/klauspost/cpuid/v2 v2.2.5 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.1 // indirect
	golang.org/x/net v0.17.0 // indirect
	gopkg.in/ini.v1 v1.67.0 // indirect
)

replace github.com/kjk/notionapi => ./..

go 1.16
