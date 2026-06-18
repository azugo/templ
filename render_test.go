package templ

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"azugo.io/azugo"
	"azugo.io/azugo/server"
	"github.com/a-h/templ"
	"github.com/go-quicktest/qt"
	"github.com/valyala/fasthttp"
)

// fakeInstr records whether the render op was observed and the error passed to
// its finish callback. It stands in for a tracing/metrics instrumenter.
type fakeInstr struct {
	mu        sync.Mutex
	observed  bool
	name      string
	finishErr error
}

func (r *fakeInstr) observe(_ context.Context, op string, args ...any) func(error) {
	if op != InstrumentationRender {
		return func(error) {}
	}

	r.mu.Lock()
	r.observed = true

	if len(args) > 0 {
		r.name, _ = args[0].(string)
	}
	r.mu.Unlock()

	return func(err error) {
		r.mu.Lock()
		r.finishErr = err
		r.mu.Unlock()
	}
}

func TestRenderError(t *testing.T) {
	base := errors.New("boom")

	// No name: the original error is returned unchanged.
	qt.Check(t, qt.Equals(renderError("", base), base))

	// Named: wrapped with the name, and errors.Is still reaches the original.
	wrapped := renderError("home.page", base)
	qt.Check(t, qt.ErrorIs(wrapped, base))
	qt.Check(t, qt.StringContains(wrapped.Error(), "home.page"))
}

func TestRenderInstrumentation(t *testing.T) {
	ok := templ.ComponentFunc(func(_ context.Context, w io.Writer) error {
		_, err := io.WriteString(w, "Hello")

		return err
	})
	fails := templ.ComponentFunc(func(_ context.Context, _ io.Writer) error {
		return errors.New("render failed")
	})

	tests := []struct {
		name      string
		input     templ.Component
		opts      []Option
		wantError bool
		wantName  string
	}{
		{name: "buffered", input: ok},
		{name: "streamed", input: ok, opts: []Option{Streaming()}},
		{name: "error", input: fails, wantError: true},
		{name: "named", input: ok, opts: []Option{Name("home.page")}, wantName: "home.page"},
		{name: "named streamed", input: ok, opts: []Option{Name("home.page"), Streaming()}, wantName: "home.page"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METRICS_ENABLED", "false")

			a, err := server.New(nil, server.Options{AppName: "Test", AppVer: "1.0.0"})
			qt.Assert(t, qt.IsNil(err))

			rec := &fakeInstr{}
			a.Instrumentation(rec.observe)

			app := azugo.NewTestApp(a)
			app.Start(t)
			defer app.Stop()

			app.Get("/", func(ctx *azugo.Context) {
				Render(ctx, tt.input, tt.opts...)
			})

			resp, err := app.TestClient().Get("/")
			defer fasthttp.ReleaseResponse(resp)
			qt.Assert(t, qt.IsNil(err))

			rec.mu.Lock()
			defer rec.mu.Unlock()

			qt.Check(t, qt.IsTrue(rec.observed))
			qt.Check(t, qt.Equals(rec.name, tt.wantName))

			if tt.wantError {
				qt.Check(t, qt.IsNotNil(rec.finishErr))
			} else {
				qt.Check(t, qt.IsNil(rec.finishErr))
			}
		})
	}
}

func TestRender(t *testing.T) {
	hello := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "Hello"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
		return nil
	})
	errorComponent := templ.ComponentFunc(func(ctx context.Context, w io.Writer) error {
		if _, err := io.WriteString(w, "Hello"); err != nil {
			t.Fatalf("failed to write string: %v", err)
		}
		return errors.New("handler error")
	})

	tests := []struct {
		name             string
		input            templ.Component
		opts             []Option
		expectedStatus   int
		expectedMIMEType string
		expectedBody     string
	}{
		{
			name:             "handlers return OK by default",
			input:            hello,
			expectedStatus:   fasthttp.StatusOK,
			expectedMIMEType: "text/html; charset=utf-8",
			expectedBody:     "Hello",
		},
		{
			name:             "handlers return OK by default",
			input:            hello,
			opts:             []Option{Streaming()},
			expectedStatus:   fasthttp.StatusOK,
			expectedMIMEType: "text/html; charset=utf-8",
			expectedBody:     "Hello",
		},
		{
			name:             "handlers return OK by default",
			input:            templ.Raw(`♠ ‘ &spades; &#8216;`),
			expectedStatus:   fasthttp.StatusOK,
			expectedMIMEType: "text/html; charset=utf-8",
			expectedBody:     "♠ ‘ &spades; &#8216;",
		},
		{
			name:             "handlers can be configured to return an alternative status code and content type",
			input:            hello,
			opts:             []Option{ContentType("text/csv")},
			expectedStatus:   fasthttp.StatusOK,
			expectedMIMEType: "text/csv",
			expectedBody:     "Hello",
		},
		{
			name:             "handlers that fail return a 500 error",
			input:            errorComponent,
			expectedStatus:   fasthttp.StatusInternalServerError,
			expectedMIMEType: "text/plain; charset=utf-8",
			expectedBody:     "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("METRICS_ENABLED", "false")
			a, err := server.New(nil, server.Options{
				AppName: "Test",
				AppVer:  "1.0.0",
			})

			qt.Assert(t, qt.IsNil(err))

			app := azugo.NewTestApp(a)
			app.Start(t)
			defer app.Stop()

			app.Get("/", func(ctx *azugo.Context) {
				Render(ctx, tt.input, tt.opts...)
			})

			resp, err := app.TestClient().Get("/")
			defer fasthttp.ReleaseResponse(resp)

			qt.Assert(t, qt.IsNil(err))
			qt.Assert(t, qt.Equals(resp.StatusCode(), tt.expectedStatus))
			qt.Assert(t, qt.Equals(string(resp.Header.ContentType()), tt.expectedMIMEType))
			qt.Assert(t, qt.Equals(string(resp.Body()), tt.expectedBody))
		})
	}
}
