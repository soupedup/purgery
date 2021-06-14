# client

Package client implements a Go client for `purgery`.

## Usage

```go
package main

import (
	"context"
	"log"

	"github.com/soupedup/purgery/pkg/client"
)

func main() {
	purgery := client.New("http://localhost:7979", "my-api-key")

	if err := purgery.Purge(context.TODO(), "http://google.com"); err != nil {
		log.Fatalf("failed purging: %v", err)
	}
}
```
