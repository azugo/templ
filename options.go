package templ

type options struct {
	Streaming   bool
	ContentType string
}

// Streaming render directly to response instead of buffering.
//
// It is recommended to use this option for large responses but can result
// in partial responses being written to the client if the component errors.
func Streaming() Option {
	return streaming(true)
}

type streaming bool

func (s streaming) apply(opt *options) {
	opt.Streaming = bool(s)
}

// ContentType sets the content type of the response.
type ContentType string

func (c ContentType) apply(opt *options) {
	opt.ContentType = string(c)
}
