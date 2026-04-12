package graphnodes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestApplyEndpointOverride(t *testing.T) {
	base := Endpoint{
		URL:               "http://service:8080/api/endpoint",
		Timeout:           5 * time.Second,
		RetriesNum:        3,
		BackoffMultiplier: 2.0,
	}

	tests := []struct {
		name     string
		endpoint Endpoint
		override *string
		wantURL  string
	}{
		{
			name:     "nil override — endpoint unchanged",
			endpoint: base,
			override: nil,
			wantURL:  "http://service:8080/api/endpoint",
		},
		{
			name:     "override replaces host:port, path preserved",
			endpoint: base,
			override: ptr("localhost:9090"),
			wantURL:  "http://localhost:9090/api/endpoint",
		},
		{
			name:     "override adds port to host-only URL",
			endpoint: Endpoint{URL: "http://service/path"},
			override: ptr("new-host:9000"),
			wantURL:  "http://new-host:9000/path",
		},
		{
			name:     "override on URL with trailing path segments",
			endpoint: Endpoint{URL: "http://old:8080/v1/foo/bar"},
			override: ptr("new:1234"),
			wantURL:  "http://new:1234/v1/foo/bar",
		},
		{
			name:     "invalid URL — endpoint unchanged",
			endpoint: Endpoint{URL: "://bad-url"},
			override: ptr("localhost:9090"),
			wantURL:  "://bad-url",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := applyEndpointOverride(tc.endpoint, tc.override)
			require.Equal(t, tc.wantURL, result.URL)
			require.Equal(t, tc.endpoint.Timeout, result.Timeout)
			require.Equal(t, tc.endpoint.RetriesNum, result.RetriesNum)
			require.Equal(t, tc.endpoint.BackoffMultiplier, result.BackoffMultiplier)
		})
	}
}

func ptr(s string) *string { return &s }
