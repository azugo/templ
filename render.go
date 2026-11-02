package templ

import (
	"fmt"

	"azugo.io/azugo"
	"azugo.io/core/http"
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

func renderError(name string, err error) error {
	if name == "" {
		return err
	}

	return fmt.Errorf("render %q: %w", name, err)
}

// Option configures rendering behaviour.
type Option interface {
	apply(opt *options)
}

// Render a component to the response.
func Render(ctx *azugo.Context, component templ.Component, opts ...Option) {
	o := &options{
		ContentType: "text/html; charset=utf-8",
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	ctx.Response().ResetBody()

	finish := observeRender(ctx, o.Name)

	err := component.Render(ctx, ctx.Response().BodyWriter())
	finish(err)

	if err != nil {
		ctx.Response().ResetBody()
		ctx.Error(renderError(o.Name, err))

		return
	}

	ctx.Header.Set(http.HeaderContentType, o.ContentType)
}
