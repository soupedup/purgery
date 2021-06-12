package purge

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/soupedup/purgery/internal/log"
)

// Func is the set of functions capable of purging the cache.
type Func func(ctx context.Context, logger *zap.Logger, varnishAddr string) bool

// New initializes and returns a Func which purges against the given Varnish
// address.
func New(addr string) Func {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 5 * time.Second,
	}

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
				return dialer.DialContext(ctx, network, addr)
			},
			ForceAttemptHTTP2:     true,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}

	return newPurgeFunc(&client)
}

func newPurgeFunc(client *http.Client) Func {
	return func(ctx context.Context, logger *zap.Logger, url string) (ok bool) {
		logger = logger.With(log.URL(url))
		logger.Info("purging ...")

		req, err := http.NewRequestWithContext(ctx, "BAN", url, nil)
		if err != nil {
			logger.Warn("failed creating purge request.",
				zap.Error(err))

			return
		}

		res, err := client.Do(req)
		if err != nil {
			logger.Warn("failed retrieving purge response.",
				zap.Error(err))

			return
		}
		defer res.Body.Close()

		switch ok = res.StatusCode == http.StatusOK; ok {
		default:
			logger.Warn("received wrong purge status code.",
				zap.Int("code", res.StatusCode))
		case true:
			logger.Debug("purged.")
		}

		return
	}
}

// errInvalidStatusCode is returned to callers of Funcs when the status code
// isn't 200.
type errInvalidStatusCode int

// Error implements error for errInvalidStatusCode
func (err errInvalidStatusCode) Error() string {
	return fmt.Sprintf("bad status code: %d", err)
}
