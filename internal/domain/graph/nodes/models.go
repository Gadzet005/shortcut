package graphnodes

import "time"

type Endpoint struct {
	URL     string
	Timeout time.Duration
}
