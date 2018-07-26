# About notionapi

This is an unofficial, read-only Go API for https://notion.so

It allows you to retrieve content of a public Notion page in structured format.

You can then convert that format to HTML.

API docs: https://godoc.org/github.com/kjk/notion

You can learn how [I reverse-engineered the API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html).

# Real-life usage

I use this API to publish my [blog](https://blog.kowalczyk.info/). Notion is my CMS where I write and edit all content and I use a custom Go script which uses this library to convert Notion pages to HTML and publish the result to Netlify.

You can see the code at https://github.com/kjk/blog

# Usage

This API only support public pages i.e. you have to share the page in Notion UI.

This is becacuse I didn't bother figuring out how to authenticate.

Then you have to know id of the page. It's the last part in Notion URL e.g. https://www.notion.so/Test-page-all-c969c9455d7c4dd79c7f860f3ace6429 has id `c969c9455d7c4dd79c7f860f3ace6429`.

Then you can retrive the page content:
```go

import (
    "log"
    "github.com/kjk/notionapi"
)

    pageID := "c969c9455d7c4dd79c7f860f3ace6429"
    page, err := notionapi.DownloadPage(pageID)
    if err != nil {
        log.Fatalf("DownloadPage() failed with %s\n", err)
    }
    // look at page.Page to see structured content
```

You can see a full example that adds recursive downloading of pages, caching etc. at https://github.com/kjk/blog/blob/master/notion_import.go

A page in notion is a tree of blocks of different types. See https://github.com/kjk/notionapi/blob/master/get_record_values.go#L31 for the definition.

To convert Notion page to HTML you can use https://github.com/kjk/blog/blob/master/notion_to_html.go as a template.
