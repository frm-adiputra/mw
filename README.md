MW
=======

[![Build Status](https://travis-ci.org/frm-adiputra/mw.svg?branch=master)](https://travis-ci.org/frm-adiputra/mw)
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ffrm-adiputra%2Fmw.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Ffrm-adiputra%2Fmw?ref=badge_shield)

MW enables [Alice](https://github.com/justinas/alice)-style chaining of context-aware middleware and handlers (using Google's `net/context` package).
But unlike [Apollo](https://github.com/cyclopsci/apollo), it never stores contexts.

```go
func (context.Context, http.ResponseWriter, *http.Request) context.Context
```

# Usage

```go
// Request flow
// mw1 -> mw2 -> mw3 -> h
handler := mw.New(mw1, mw2, mw3).Then(h)

// Store a chain and use later
c0 := mw.New(mw1, mw2, mw3)
c1 := c0.Chain(mw4, mw5)

// mw1 -> mw2 -> mw3 -> h
handler = c0.Then(h)

// mw1 -> mw2 -> mw3 -> mw4 -> mw5 -> h
handler = c1.Then(h)
```

# Integration with http.Handler middleware

MW provides a `Wrap` function to inject normal http.Handler-based middleware into the chain.
The context will skip over the injected middleware and pass unharmed to the next context-aware handler in the chain.

```go
mw.New(mw1, mw.Wrap(NormalMiddlware), mw2).Then(handler)
```


## License
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Ffrm-adiputra%2Fmw.svg?type=large)](https://app.fossa.io/projects/git%2Bgithub.com%2Ffrm-adiputra%2Fmw?ref=badge_large)