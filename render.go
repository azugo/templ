package templ

import (
	"bytes"
	"sync"

	"azugo.io/azugo"
	"github.com/a-h/templ"
)

// InstrumentationRender is the instrumentation operation name emitted around a
// component render. When a name is provided via the Name option it is passed as
// the first (string) argument of the event.
const InstrumentationRender = "templ-render"

func observeRender(ctx *azugo.Context, name string) func(error) {
	inst := ctx.App().Instrumenter()
	if name == "" || inst.Empty() {
		return inst.Observe(ctx, InstrumentationRender)
	}

	return inst.Observe(ctx, InstrumentationRender, name)
}

// Option configures rendering behaviour.
type Option interface {
	apply(opt *options)
}

var bufferPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func renderBuffered(ctx *azugo.Context, component templ.Component, contentType, name string) {
	buf, _ := bufferPool.Get().(*bytes.Buffer)

	defer func() {
		buf.Reset()
		bufferPool.Put(buf)
	}()

	finish := observeRender(ctx, name)

	err := component.Render(ctx, buf)
	finish(err)

	if err != nil {
		ctx.Error(err)

		return
	}

	ctx.Header.Set("Content-Type", contentType)
	ctx.Raw(buf.Bytes())
}

func renderStreamed(ctx *azugo.Context, component templ.Component, contentType, name string) {
	ctx.Header.Set("Content-Type", contentType)

	wr := ctx.Context().Response.BodyWriter()

	finish := observeRender(ctx, name)

	err := component.Render(ctx, wr)
	finish(err)

	if err != nil {
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
		renderStreamed(ctx, component, o.ContentType, o.Name)

		return
	}

	renderBuffered(ctx, component, o.ContentType, o.Name)
}
