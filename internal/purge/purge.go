package purge

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const purgeURL = "http://localhost/purge"

// Node purges the node at the given URL for the given stamp.
func Node(ctx context.Context, prefix string, stamp time.Time) (err error) {
	var req *http.Request
	if req, err = http.NewRequestWithContext(ctx, "PURGE", purgeURL, nil); err != nil {
		return
	}

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
