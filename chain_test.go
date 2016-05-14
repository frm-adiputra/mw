package mwchain

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/net/context"

	"github.com/stretchr/testify/assert"
)

var fxMW = make(map[string]func(Handler) Handler)

func init() {
	for i := 0; i <= 5; i++ {
		b := fmt.Sprintf("b%d-", i)
		a := fmt.Sprintf("a%d-", i)
		f := func(h Handler) Handler {
			return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
				w.Write([]byte(b))
				c := h.ServeHTTP(ctx, w, r)
				w.Write([]byte(a))
				return c
			})
		}
		fxMW[fmt.Sprintf("mw%d", i)] = f
	}

	for i := 0; i <= 5; i++ {
		b := fmt.Sprintf("b%d-", i)
		a := fmt.Sprintf("a%d-", i)
		f := func(h Handler) Handler {
			return HandlerFunc(func(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
				var s string
				ctxVal := ctx.Value("test")
				if ctxVal != nil {
					s = ctxVal.(string)
				}

				newC := context.WithValue(ctx, "test", s+b)
				w.Write([]byte(b))

				c := h.ServeHTTP(newC, w, r)

				w.Write([]byte(a))
				ctxVal = c.Value("test")
				if ctxVal != nil {
					s = ctxVal.(string)
				}
				return context.WithValue(ctx, "test", s+a)
			})
		}
		fxMW[fmt.Sprintf("mwctx%d", i)] = f
	}
}

func hZero(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	w.Write([]byte("h0-"))
	return ctx
}

func hZeroCtx(ctx context.Context, w http.ResponseWriter, r *http.Request) context.Context {
	w.Write([]byte("h0-"))

	var s string
	ctxVal := ctx.Value("test")
	if ctxVal != nil {
		s = ctxVal.(string)
	}

	return context.WithValue(ctx, "test", s+"h0-")
}

func httpMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hxb0-"))
		h.ServeHTTP(w, r)
		w.Write([]byte("hxa0-"))
	})
}

func TestChain(t *testing.T) {
	assert := assert.New(t)
	chain := New(
		fxMW["mw0"],
		fxMW["mw1"],
		fxMW["mw2"],
	).Then(HandlerFunc(hZero))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chain.ServeHTTP(context.Background(), w, r)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(err)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal(string(body), "b0-b1-b2-h0-a2-a1-a0-")
}

func TestChainWithNilAsFinal(t *testing.T) {
	assert := assert.New(t)
	chain := New(
		fxMW["mw0"],
		fxMW["mw1"],
		fxMW["mw2"],
	).Then(nil)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chain.ServeHTTP(context.Background(), w, r)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(err)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal(string(body), "b0-b1-b2-a2-a1-a0-")
}

func TestChainNoMW(t *testing.T) {
	assert := assert.New(t)
	chain := New().Then(HandlerFunc(hZero))
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chain.ServeHTTP(context.Background(), w, r)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(err)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal(string(body), "h0-")
}

func TestChainNoMWWithNilAsFinal(t *testing.T) {
	assert := assert.New(t)
	chain := New().Then(nil)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chain.ServeHTTP(context.Background(), w, r)
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(err)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal(string(body), "")
}

func TestChainAppend(t *testing.T) {
	assert := assert.New(t)
	chain0 := New(
		fxMW["mw0"],
		fxMW["mw1"],
		fxMW["mw2"],
	)

	mwh1 := chain0.Chain(
		fxMW["mw3"],
		fxMW["mw4"],
		fxMW["mw5"],
	).Then(HandlerFunc(hZero))

	mwh0 := chain0.Then(HandlerFunc(hZero))

	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mwh0.ServeHTTP(context.Background(), w, r)
	}))
	defer ts0.Close()

	res, err := http.Get(ts0.URL)
	assert.NoError(err)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal(string(body), "b0-b1-b2-h0-a2-a1-a0-")

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mwh1.ServeHTTP(context.Background(), w, r)
	}))
	defer ts1.Close()

	res, err = http.Get(ts1.URL)
	assert.NoError(err)

	body, err = ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal(string(body), "b0-b1-b2-b3-b4-b5-h0-a5-a4-a3-a2-a1-a0-")
}

func TestChainUseContext(t *testing.T) {
	assert := assert.New(t)
	chain := New(
		fxMW["mwctx0"],
		fxMW["mwctx1"],
		fxMW["mwctx2"],
	).Then(HandlerFunc(hZeroCtx))
	var s string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := chain.ServeHTTP(context.Background(), w, r)
		v := c.Value("test")
		if v != nil {
			s = v.(string)
		}
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal("b0-b1-b2-h0-a2-a1-a0-", s)
}

func TestChainWrapUseContext(t *testing.T) {
	assert := assert.New(t)
	chain := New(
		fxMW["mwctx0"],
		fxMW["mwctx1"],
		Wrap(httpMW),
		fxMW["mwctx2"],
	).Then(HandlerFunc(hZeroCtx))
	var s string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := chain.ServeHTTP(context.Background(), w, r)
		v := c.Value("test")
		if v != nil {
			s = v.(string)
		}
	}))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(err)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal("b0-b1-hxb0-b2-h0-a2-hxa0-a1-a0-", string(body))

	assert.Equal(res.StatusCode, 200)
	assert.Equal("b0-b1-b2-h0-a2-a1-a0-", s)
}
