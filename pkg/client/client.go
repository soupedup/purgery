// Package client implements a purgery client.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
	"sync"
)

// New initializes and returns a Purgery client for the purgery API running at
// the given API root URL.
//
// In case the given apiKey isn't empty, the client will use HTTP Basic
// Authorization when interacting with the API.
func New(rootURL, apiKey string) *Client {
	for {
		if strings.HasSuffix(rootURL, "/") {
			rootURL = rootURL[:len(rootURL)-1]

			continue
		}

		break
	}

	return &Client{
		http:     buildClient(apiKey),
		rootURL:  rootURL,
		purgeURL: joinURL(rootURL, "purge"),
	}
}

func joinURL(root string, segments ...string) string {
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}

	return root + path.Join(segments...)
}

// Client implements a Purgery client.
//
// Instances of Client are safe for concurrent use.
type Client struct {
	http     *http.Client
	rootURL  string
	purgeURL string
}

// Root returns the root URL of the API the Client is configured to run against.
func (c *Client) RootURL() string { return c.rootURL }

type errInvalidStatusCode int

func (err errInvalidStatusCode) Error() string {
	return fmt.Sprintf("purgery: invalid response status code (%d)", int(err))
}

var errInternalServerError = errors.New("purgery: internal server error")

type errInvalidURL string

func (err errInvalidURL) Error() string {
	return fmt.Sprintf("purgery: invalid url (%q)", string(err))
}

// Purge requests that the given URL be purged from the remote cache.
func (c *Client) Purge(ctx context.Context, url string) (err error) {
	payload := struct {
		URL string `json:"url"`
	}{
		URL: url,
	}

	enc := checkoutEncoder()
	defer enc.release()

	if err = enc.Encode(payload); err != nil {
		return
	}

	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.purgeURL, enc); err != nil {
		return
	}

	var res *http.Response
	if res, err = c.http.Do(req); err != nil {
		return
	}
	res.Body.Close()

	switch res.StatusCode {
	default:
		err = errInvalidStatusCode(res.StatusCode)
	case http.StatusNoContent:
		break
	case http.StatusUnprocessableEntity:
		err = errInvalidURL(url)
	case http.StatusInternalServerError:
		err = errInternalServerError
	}

	return
}

func buildClient(apiKey string) *http.Client {
	return &http.Client{
		Transport: &apiKeyAuth{
			RoundTripper: http.DefaultTransport,
			apiKey:       apiKey,
		},
	}
}

type apiKeyAuth struct {
	http.RoundTripper
	apiKey string
}

func (aka *apiKeyAuth) RoundTrip(req *http.Request) (*http.Response, error) {
	if aka.apiKey != "" {
		req.SetBasicAuth(aka.apiKey, "")
	}

	return aka.RoundTripper.RoundTrip(req)
}

type encoder struct {
	*json.Encoder
	*bytes.Buffer
}

func (e *encoder) release() {
	e.Buffer.Reset()

	encoders.Put(e)
}

var encoders = sync.Pool{
	New: func() interface{} {
		buf := new(bytes.Buffer)

		return &encoder{
			Encoder: json.NewEncoder(buf),
			Buffer:  buf,
		}
	},
}

func checkoutEncoder() *encoder {
	return encoders.Get().(*encoder)
}
