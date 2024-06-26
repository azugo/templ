# Azugo Templ rendering

[![status-badge](https://ci.azugo.io/api/badges/azugo/templ/status.svg)](https://ci.azugo.io/azugo/templ)

Azugo framework [Templ](https://templ.guide/) HTML template rendering support.

## Usage

```go
	app.Get("/", func(ctx *azugo.Context) {
		templ.Render(ctx, index())
	})
```
