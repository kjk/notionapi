# About notionapi

This is an unofficial, Go API for https://notion.so. Mostly for reading, limited write capabilities.

It allows you to retrieve content of a Notion page in structured format.

You can then e.g. convert that format to HTML.

API docs: https://godoc.org/github.com/kjk/notionapi

Tutorial: https://www.programming-books.io/essential/go/db2797b0772a42fca5820014164589a7

You can learn how [I reverse-engineered the Notion API](https://blog.kowalczyk.info/article/88aee8f43620471aa9dbcad28368174c/how-i-reverse-engineered-notion-api.html) in order to write this library.

# Real-life usage

I use this API to publish my [blog](https://blog.kowalczyk.info/) and series of [programming books](https://www.programming-books.io/) from content stored in Notion.

Notion serves as a CMS (Content Management System). I write and edit pages in Notion.

I use custom Go program to download Notion pages using this this library and converts pages to HTML. It then publishes the result to Netlify (but it could be )

You can see the code at https://github.com/kjk/blog and https://github.com/essentialbooks/books/tree/master/cmd/gen-books

# Implementations for other languages

* https://github.com/jamalex/notion-py : library for Python

