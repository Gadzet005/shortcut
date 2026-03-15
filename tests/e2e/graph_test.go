package e2e

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	baseURL = "http://localhost:8080"
)

func TestGraphRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping e2e test in short mode")
	}

	tests := []struct {
		name           string
		graphID        string
		input          string
		expectedOutput string
		expectedStatus int
	}{
		{
			name:           "successful graph execution with number 3244",
			graphID:        "sum-echoes",
			input:          "3244",
			expectedOutput: "6488",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful graph execution with number 100",
			graphID:        "sum-echoes",
			input:          "100",
			expectedOutput: "200",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful graph execution with number 0",
			graphID:        "sum-echoes",
			input:          "0",
			expectedOutput: "0",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "successful graph execution with negative number",
			graphID:        "sum-echoes",
			input:          "-50",
			expectedOutput: "-100",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := resty.New()
			resp, err := client.R().
				SetBody(tt.input).
				Post(baseURL + "/graph/" + tt.graphID + "/run")
			require.NoError(t, err, "failed to execute request")

			assert.Equal(t, tt.expectedStatus, resp.StatusCode(), "unexpected status code")
			assert.Equal(t, tt.expectedOutput, resp.String(), "unexpected result")
		})
	}
}
