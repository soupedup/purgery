package purge

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
)

var purgeURL = os.Getenv("PURGE_URL")

// URL purges the given URL from the local node.
func URL(ctx context.Context, url string) (err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, "PURGE", purgeURL, strings.NewReader(url)); err != nil {
		return
	}

	// TODO(@azazeal): implement timeouts and backoff

	var res *http.Response
	if res, err = http.DefaultClient.Do(req); err != nil {
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
