# About notionapi

This is an unofficial, Go API for https://notion.so. Mostly for reading, limited write capabilities.

It allows you to retrieve content of a Notion page in structured format.

You can then e.g. convert that format to HTML.

API docs: https://godoc.org/github.com/kjk/notionapi

You can learn how [I reverse-engineered the API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html).

# Real-life usage

I use this API to publish my [blog](https://blog.kowalczyk.info/) and series of [programming books](https://www.programming-books.io/). 

Notion is my CMS where I write and edit all content. I use custom Go program which uses this library to convert Notion pages to HTML and publish the result to Netlify.

You can see the code at https://github.com/kjk/blog

# Usage

Then you have to know id of the page. It's the last part in Notion URL e.g. https://www.notion.so/Test-page-all-c969c9455d7c4dd79c7f860f3ace6429 has id `c969c9455d7c4dd79c7f860f3ace6429`.

Then you can retrive the content of public page:
```go

import (
    "log"
    "github.com/kjk/notionapi"
)

    client := &notionapi.Client{}
    pageID := "c969c9455d7c4dd79c7f860f3ace6429"
    page, err := client.DownloadPage(pageID)
    if err != nil {
        log.Fatalf("DownloadPage() failed with %s\n", err)
    }
    // look at page.Page to see structured content
```


# Accessing non-public pages

To access non-public pages you need to find out authentication token. 

Auth token is the value of `token_v2` cookie. 

In Chrome: open developer tools (Menu `More Tools\Developer Tools`), navigate to `Application` tab, look under `Storage \ Cookies` and copy the value of `token_v2` cookie. You can do similar things in other browsers.

Then configure `Client` with access token::
```
client := &notionapi.Client{}
client.AuthToken = "value of token_v2 value"
```

# Writing data to Notion

Currently the library has very limited capabilities for writing data to Notion.

You can change page title and change format of the page.

For example, to change page title:

```go
page, err := client.DownloadPage(pageID)
if err != nil {
    log.Fatalf("DownloadPage() failed with %s\n", err)
}
err = page.SetTitle("new title")
if err != nil {
    log.Fatalf("SetTitle() failed with %s\n", err)
}
```

There's no technical reason write capabilities are so limited. We could do everything that Notion does. I just don't need such capability myself.

If you have specific needs, open an issue and describe what write capabilities you need and I'll look into it.

# Examples

You can see a full example that adds recursive downloading of pages, caching etc. at https://github.com/kjk/blog/blob/master/notion_import.go

A page in notion is a tree of blocks of different types. See https://github.com/kjk/notionapi/blob/master/get_record_values.go#L31 for the definition.

To convert Notion page to HTML you can use https://github.com/kjk/blog/blob/master/notion_to_html.go as a template.

# Implementations for other languages

* https://github.com/jamalex/notion-py : for Python, even has more functionality
