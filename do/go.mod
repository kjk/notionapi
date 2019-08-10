module github.com/kjk/notionapi/do

go 1.12

require (
	github.com/kjk/caching_http_client v0.0.0-20190810075619-06ff809674f7
	github.com/kjk/notionapi v0.0.0
	github.com/kjk/siser v0.0.0-20190801014033-b3367920d7f2
	github.com/yosssi/gohtml v0.0.0-20190128141317-9b7db94d32d9
)

replace github.com/kjk/notionapi => ./..
