package graphnodes

import (
	"net/url"
	"time"
)

type Endpoint struct {
	URL               string
	Timeout           time.Duration
	RetriesNum        int
	InitialInterval   time.Duration
	BackoffMultiplier float64
	MaxInterval       time.Duration
}

func applyEndpointOverride(endpoint Endpoint, override *string) Endpoint {
	if override == nil {
		return endpoint
	}
	u, err := url.Parse(endpoint.URL)
	if err != nil {
		return endpoint
	}
	u.Host = *override
	endpoint.URL = u.String()
	return endpoint
}
