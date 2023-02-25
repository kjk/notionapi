module github.com/kjk/notionapi/do

require (
	github.com/kjk/atomicfile v0.0.0-20220410204726-989ae30d2b66 // indirect
	github.com/kjk/fmthtml v0.0.0-20190816041536-39f5e479d32d
	github.com/kjk/notionapi v0.0.0-20220629202131-bb7dc156793e
	github.com/kjk/u v0.0.0-20220410204605-ce4a95db4475
	github.com/klauspost/cpuid/v2 v2.0.14 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/net v0.7.0 // indirect
	gopkg.in/ini.v1 v1.66.6 // indirect
)

replace github.com/kjk/notionapi => ./..

go 1.16
