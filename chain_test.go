// mwchain provides `net/context`-aware middleware chaining
package mwchain

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultNew(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		chain := New()
		assert.Len(chain.constructors, 0)
	})
}

func TestThenNoMiddleware(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		chain := New()
		final := chain.Then(HandlerFunc(handlerOne))
		assert.Implements((*Handler)(nil), final)
	})
}

func TestThenPureNoMiddleware(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		chain := New()
		final := chain.ThenPure(HandlerFunc(handlerOne))
		assert.Implements((*http.Handler)(nil), final)
	})
}

func TestThenFuncNoMiddleware(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		chain := New()
		final := chain.ThenFunc(handlerOne)
		assert.Implements((*Handler)(nil), final)
	})
}

func TestThenPureFuncNoMiddleware(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		chain := New()
		final := chain.ThenPureFunc(handlerOne)
		assert.Implements((*http.Handler)(nil), final)
	})
}

func TestThenNil(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		final := New().Then(nil)
		assert.Implements((*Handler)(nil), final)
	})
}

func TestThenPureNil(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		final := New().ThenPure(nil)
		assert.Implements((*http.Handler)(nil), final)
	})
}

func TestThenFuncNil(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		final := New().ThenFunc(nil)
		assert.Implements((*Handler)(nil), final)
	})
}

func TestThenPureFuncNil(t *testing.T) {
	assert := assert.New(t)
	assert.NotPanics(func() {
		final := New().ThenPureFunc(nil)
		assert.Implements((*http.Handler)(nil), final)
	})
}

func TestAppend(t *testing.T) {
	assert := assert.New(t)
	chain := New(middleOne)
	newChain := chain.Append(middleTwo)
	assert.Len(chain.constructors, 1)
	assert.Len(newChain.constructors, 2)
}

func TestChains(t *testing.T) {
	assert := assert.New(t)
	setCtx := NewTestContext(10)

	chain := New(setCtx, middleOne, middleTwo).ThenPureFunc(handlerContext)

	ts := httptest.NewServer(chain)
	defer ts.Close()

	res, err := http.Get(ts.URL)
	assert.NoError(err)

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(err)
	res.Body.Close()

	assert.Equal(res.StatusCode, 200)
	assert.Equal(string(body), "m1\nm2\n10\n")
}
