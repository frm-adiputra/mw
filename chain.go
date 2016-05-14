package mwchain

import (
	"net/http"

	"golang.org/x/net/context"
)

var noopHandler = HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return ctx
})

// Handler is a context-aware interface analagous to the `net/http` http.Handler
// interface. The differences are that a context.Context is required as the
// first parameter in ServeHTTP, and it also returns a context.
type Handler interface {
	ServeHTTP(context.Context, http.ResponseWriter, *http.Request) context.Context
}

// HandlerFunc similar to http.HandlerFunc, is an adapter to convert functions
// into Handler interface.
type HandlerFunc func(context.Context, http.ResponseWriter, *http.Request) context.Context

// ServeHTTP calls the wrapped function h(ctx, w, r)
func (h HandlerFunc) ServeHTTP(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	return h(ctx, w, r)
}

// Constructor is a constructor for a context-aware middleware.
type Constructor func(Handler) Handler

// Chain acts as a list of mwchain.Handler constructors. Chain is effectively
// immutable, once created, it will always hold the same set of constructors in
// the same order.
type Chain struct {
	constructors []Constructor
}

// NewChain creates a new chain of middlewares. Each constructors are only
// called upon a call to Then().
func NewChain(constructors ...Constructor) Chain {
	c := Chain{}
	c.constructors = append(c.constructors, constructors...)

	return c
}

// Then chains the middleware and returns the final Handler.
func (c Chain) Then(last Handler) Handler {
	_last := last

	if _last == nil {
		_last = noopHandler
	}

	for i := len(c.constructors) - 1; i >= 0; i-- {
		_last = c.constructors[i](_last)
	}

	return _last
}

// Append extends a chain, adding the specified constructors
// as the end of the chain.
//
// Append returns a new chain, leaving the original one untouched.
func (c Chain) Append(constructors ...Constructor) Chain {
	newCons := make([]Constructor, len(c.constructors)+len(constructors))
	copy(newCons, c.constructors)
	copy(newCons[len(c.constructors):], constructors)

	return NewChain(newCons...)
}

// Wrap allows injection of normal http.Handler middleware into a chain. The
// context will be preserved and passed through intact
func Wrap(h func(http.Handler) http.Handler) Constructor {

	return func(next Handler) Handler {

		return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
			var _ctx context.Context
			hNext := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_ctx = next.ServeHTTP(ctx, w, r)
			})

			result := h(hNext)
			result.ServeHTTP(w, r)
			return _ctx
		})
	}
}
