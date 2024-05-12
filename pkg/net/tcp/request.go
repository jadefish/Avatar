package tcp

import (
	"context"
	"io"
)

// A Request received by a server or to be sent by a client.
type Request struct {
	// Body is the request's payload. It is always non-nil.
	Body io.ReadCloser

	// Length is the size, in bytes, of the request's payload.
	Length int16

	// Close indicates whether to close the connection after replying to the
	// request.
	Close bool

	// RemoteAddr is the network address that sent the request. It has no
	// defined format. The TCP server in this package sets RemoteAddr to an
	// "IP:port" address before invoking a handler.
	RemoteAddr string

	ctx context.Context
}

// Context returns the request's context. To change the context, use
// WithContext.
//
// The returned context is always non-nil; it defaults to the
// background context.
//
// For outgoing client requests, the context controls cancellation.
//
// For incoming server requests, the context is canceled when the
// client's connection closes, the request is canceled (with HTTP/2),
// or when the ServeHTTP method returns.

func (r *Request) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}

	return context.Background()
}
