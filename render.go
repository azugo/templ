package templ

import (
	"bytes"
	"sync"

	"azugo.io/azugo"
	"github.com/a-h/templ"
)

type Option interface {
	apply(opt *options)
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func renderBuffered(ctx *azugo.Context, component templ.Component, contentType string) {
	buf, _ := bufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	if err := component.Render(ctx, buf); err != nil {
		ctx.Error(err)

		return
	}

	ctx.Header.Set("Content-Type", contentType)
	ctx.Raw(buf.Bytes())
}

func renderStreamed(ctx *azugo.Context, component templ.Component, contentType string) {
	ctx.Header.Set("Content-Type", contentType)

	wr := ctx.Context().Response.BodyWriter()

	if err := component.Render(ctx, wr); err != nil {
		ctx.Error(err)
	}
}

// Render a component to the response.
func Render(ctx *azugo.Context, component templ.Component, opts ...Option) {
	o := &options{
		ContentType: "text/html; charset=utf-8",
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	if o.Streaming {
		renderStreamed(ctx, component, o.ContentType)

		return
	}

	renderBuffered(ctx, component, o.ContentType)
}
