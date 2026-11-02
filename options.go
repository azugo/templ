// Package templ provides helpers for rendering templ components in Azugo applications.
package templ

type options struct {
	ContentType string
	Name        string
}

// ContentType sets the content type of the response.
type ContentType string

func (c ContentType) apply(opt *options) {
	opt.ContentType = string(c)
}

// Name of the component, used in instrumentation (e.g. the tracing span name)
// and to annotate render errors so the failing template is identifiable.
type Name string

func (n Name) apply(opt *options) {
	opt.Name = string(n)
}
