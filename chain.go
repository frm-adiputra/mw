// Package mwchain provides `net/context`-aware middleware chaining
package mwchain

import (
	"net/http"

	"golang.org/x/net/context"
)

// Constructor is a constructor for a piece of context-aware middleware.
type Constructor func(Handler) Handler

// Chain acts as a list of mwchain.Handler constructors.
// Chain is effectively immutable:
// once created, it will always hold
// the same set of constructors in the same order.
// Chain also holds a copy of the context to be injected
// to the first middleware when .Then() is called.
type Chain struct {
	constructors []Constructor
}

// New creates a new chain,
// memorizing the given list of middleware constructors.
// New serves no other function,
// constructors are only called upon a call to Then().
func New(constructors ...Constructor) Chain {
	c := Chain{}
	c.constructors = append(c.constructors, constructors...)

	return c
}

// ThenPure chains the middleware and returns the final http.Handler.
//     New(m1, m2, m3).ThenPure(h)
// is equivalent to:
//     m1(m2(m3(h)))
// When the request comes in, it will be passed to m1, then m2, then m3
// and finally, the given handler
// (assuming every middleware calls the following one).
//
// A chain can be safely reused by calling ThenPure() several times.
//     stdStack := alice.New(ratelimitHandler, csrfHandler)
//     indexPipe = stdStack.ThenPure(indexHandler)
//     authPipe = stdStack.ThenPure(authHandler)
// Note that constructors are called on every call to ThenPure()
// and thus several instances of the same middleware will be created
// when a chain is reused in this way.
// For proper middleware, this should cause no problems.
//
// ThenPure() treats nil as http.DefaultServeMux.
func (c Chain) ThenPure(h Handler) http.Handler {
	var final Handler
	if h != nil {
		final = h
	} else {
		final = &stripsContext{
			handler: http.DefaultServeMux,
		}
	}

	for i := len(c.constructors) - 1; i >= 0; i-- {
		final = c.constructors[i](final)
	}

	adapter := addsContext{
		ctx:     context.Background(),
		handler: final,
	}

	return &adapter
}

// Then chains the middleware and returns the final Handler.
func (c Chain) Then(h Handler) Handler {
	var final Handler
	if h != nil {
		final = h
	} else {
		final = &noopHandler{}
	}

	for i := len(c.constructors) - 1; i >= 0; i-- {
		final = c.constructors[i](final)
	}

	adapter := ctxHandler{
		handler: final,
	}

	return &adapter
}

// ThenFunc works identically to Then, but takes
// a HandlerFunc instead of a Handler.
//
// The following two statements are equivalent:
//     c.ThenPure(http.HandlerFunc(fn))
//     c.ThenFunc(fn)
//
// ThenFunc provides all the guarantees of Then.
func (c Chain) ThenFunc(fn func(context.Context, http.ResponseWriter, *http.Request)) Handler {
	if fn == nil {
		return c.Then(nil)
	}
	return c.Then(HandlerFunc(fn))
}

// ThenPureFunc provides all the guarantees of ThenPure.
func (c Chain) ThenPureFunc(fn HandlerFunc) http.Handler {
	if fn == nil {
		return c.ThenPure(nil)
	}
	return c.ThenPure(HandlerFunc(fn))
}

// Append extends a chain, adding the specified constructors
// as the last ones in the request flow.
//
// Append returns a new chain, leaving the original one untouched.
//
//     stdChain := alice.New(m1, m2)
//     extChain := stdChain.Append(m3, m4)
//     // requests in stdChain go m1 -> m2
//     // requests in extChain go m1 -> m2 -> m3 -> m4
func (c Chain) Append(constructors ...Constructor) Chain {
	newCons := make([]Constructor, len(c.constructors)+len(constructors))
	copy(newCons, c.constructors)
	copy(newCons[len(c.constructors):], constructors)

	return New(newCons...)
}
