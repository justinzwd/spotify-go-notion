Spotify And Notion API

# Installation

```
$ go get github.com/jomei/notionapi
```

# Getting started
Follow Notion’s [getting started guide](https://developers.notion.com/docs/getting-started) to obtain an Integration Token.

## Example

Make a new `Client`

```go
import "github.com/jomei/notionapi"


client := notionapi.NewClient("your-integration-token")
```
Then, use client's methods to retrieve or update your content

```go
page, err := client.Page.Get(context.Background(), "your-page-id")
if err != nil {
	// do something
}
```
