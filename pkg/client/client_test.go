package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPurge(t *testing.T) {
	srv := newServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			panic(fmt.Errorf("invalid method: %q", r.Method))
		}

		var payload struct {
			URL string `json:"url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			panic(err)
		}

		status := func(code int) {
			http.Error(w, http.StatusText(code), code)
		}

		switch payload.URL {
		case "ok":
			w.WriteHeader(http.StatusNoContent)
		case "invalidStatusCode":
			status(http.StatusTeapot)
		case "unprocessableEntity":
			status(http.StatusUnprocessableEntity)
		case "internalServerError":
			status(http.StatusInternalServerError)
		}
	})
	defer srv.Close()

	client := New(srv.URL+"//", "123")
	assert.Equal(t, srv.URL, client.RootURL())
	assert.Equal(t, time.Minute>>1, client.http.Timeout)

	cases := map[string]struct {
		ctx context.Context
		exp error
	}{
		"nil": {
			exp: errors.New("net/http: nil Context"),
		},
		"ok": {
			ctx: context.Background(),
		},
		"cancelled": {
			ctx: newCancelledContext(),
			exp: &url.Error{
				Op:  "Post",
				URL: srv.URL + "/purge",
				Err: context.Canceled,
			},
		},
		"timeout": {
			ctx: newTimedOutContext(),
			exp: &url.Error{
				Op:  "Post",
				URL: srv.URL + "/purge",
				Err: context.DeadlineExceeded,
			},
		},
		"invalidStatusCode": {
			ctx: context.Background(),
			exp: errInvalidStatusCode(http.StatusTeapot),
		},
		"unprocessableEntity": {
			ctx: context.Background(),
			exp: errInvalidURL("unprocessableEntity"),
		},
		"internalServerError": {
			ctx: context.Background(),
			exp: errInternalServerError,
		},
	}

	for url := range cases {
		kase := cases[url]

		got := client.Purge(kase.ctx, url)
		t.Log(got)
		assert.Equal(t, kase.exp, got, "url: %s", url)
	}
}

func newServer(fn http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(fn))
}

func newCancelledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	return ctx
}

func newTimedOutContext() context.Context {
	ctx, cancel := context.WithDeadline(context.Background(), time.Now())
	cancel()

	return ctx
}

func TestErrInvalidURLError(t *testing.T) {
	fn := func(url string) bool {
		return errInvalidURL(url).Error() ==
			fmt.Sprintf("purgery: invalid url (%q)", url)
	}

	require.NoError(t, quick.Check(fn, nil))
}

func TestErrInvalidStatusCode(t *testing.T) {
	fn := func(code int) bool {
		return errInvalidStatusCode(code).Error() ==
			fmt.Sprintf("purgery: invalid response status code (%d)", code)
	}

	require.NoError(t, quick.Check(fn, nil))
}
