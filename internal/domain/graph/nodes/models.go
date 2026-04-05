package graphnodes

import "time"

type Endpoint struct {
	URL               string
	Timeout           time.Duration
	RetriesNum        int
	InitialInterval   time.Duration
	BackoffMultiplier float64
	MaxInterval       time.Duration
}
