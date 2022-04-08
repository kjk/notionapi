module github.com/kjk/notionapi/do

require (
	github.com/kjk/atomicfile v0.0.0-20210818091506-2c406a58bae3 // indirect
	github.com/kjk/fmthtml v0.0.0-20190816041536-39f5e479d32d
	github.com/kjk/notionapi v0.0.0-20220408051726-4dd3ce62c5e9
	github.com/kjk/u v0.0.0-20210327060556-13ea33918991
	github.com/klauspost/cpuid/v2 v2.0.12 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/minio/sha256-simd v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20220331220935-ae2d96664a29 // indirect
	golang.org/x/net v0.0.0-20220407224826-aac1ed45d8e3 // indirect
	golang.org/x/sys v0.0.0-20220406163625-3f8b81556e12 // indirect
	gopkg.in/ini.v1 v1.66.4 // indirect
)

replace github.com/kjk/notionapi => ./..

go 1.16
