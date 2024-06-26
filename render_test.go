package templ

import (
	"context"
	"errors"
	"io"
	"testing"

	"azugo.io/azugo"
	"azugo.io/azugo/server"
	"github.com/a-h/templ"
	"github.com/go-quicktest/qt"
	"github.com/valyala/fasthttp"
)

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
