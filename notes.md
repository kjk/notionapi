This is how "removing from bookmarks" operation looks like:

```json
{
  "operations": [
    {
      "table": "space_view",
      "id": "4e548900-40d1-4140-a0a8-165f0c373f6d",
      "path": [
        "bookmarked_pages"
      ],
      "command": "listRemove",
      "args": {
        "id": "7e825831-be07-487e-87e7-56e52914233b"
      }
    }
  ]
}
```

Operation for updating "last edited". Looks like they are sent in pairs:
for block and parent page block:
```json
{
  "id":"c969c945-5d7c-4dd7-9c7f-860f3ace6429",
  "table":"block",
  "path":[],
  "command":"update",
  "args":{
    "last_edited_time":1551762900000
  }
}
```

Operation for changing page format:
```json
{
    "id": "c969c945-5d7c-4dd7-9c7f-860f3ace6429",
    "table": "block",
    "path": [
        "format"
    ],
    "command": "update",
    "args": {
        "page_small_text": true
    }
}
```

Operation for changing language in a code block:
```json
{
    "id": "e802296a-b0dc-41a8-8aa3-cf4212c3da0b",
    "table": "block",
    "path": [ "properties" ],
    "command": "update",
    "args": {
        "language": [
            [
                "Go"
            ]
        ]
    }
}
```
