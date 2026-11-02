# Azugo Templ rendering

[![status-badge](https://ci.azugo.io/api/badges/azugo/templ/status.svg)](https://ci.azugo.io/azugo/templ)

Azugo framework [Templ](https://templ.guide/) HTML template rendering support.

## Usage

```go
	app.Get("/", func(ctx *azugo.Context) {
		templ.Render(ctx, index())
	})
```

### Fragments

Render only the `@templ.Fragment` blocks with the given ids, e.g. for htmx partial updates:

```go
	app.Get("/items", func(ctx *azugo.Context) {
		if ctx.Header.Get("HX-Request") != "" {
			templ.Render(ctx, itemsPage(), templ.Fragments("list"))

			return
		}

		templ.Render(ctx, itemsPage())
	})
```

```templ
templ itemsPage() {
	<main>
		@templ.Fragment("list") {
			<ul>...</ul>
		}
	</main>
}
```
