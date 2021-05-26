package purge

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/soupedup/purgery/internal/env"
)

var purgeAddr string

func init() {
	var err error
	if purgeAddr, err = env.PurgeAddr(); err != nil {
		panic(err)
	}
}

var (
	dialer = (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	})

	client = http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				log.Printf("dialing %q ...", purgeAddr)

				return dialer.DialContext(ctx, network, purgeAddr)
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
)

// URL purges the given URL from the local node.
func URL(ctx context.Context, url string) (err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, "PURGE", url, nil); err != nil {
		return
	}

	// TODO(@azazeal): implement timeouts and backoff

	var res *http.Response
	if res, err = client.Do(req); err != nil {
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err = errInvalidStatusCode(res.StatusCode)
	}

	return
}

type errInvalidStatusCode int

// Error implements error for errInvalidStatusCode
func (err errInvalidStatusCode) Error() string {
	return fmt.Sprintf("bad status code: %d", err)
}
