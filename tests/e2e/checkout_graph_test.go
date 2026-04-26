package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Multi-level DAG:
// input ─► parse-request ─► validate-user ──────────────────────────────┐
//
//	└─► fetch-product ─► check-inventory ────────────┼─► build-summary
//	                 └─► apply-discount ─────────────┘
func TestCheckoutSummary(t *testing.T) {
	type args struct {
		userID    string
		productID string
	}

	testCases := []struct {
		name       string
		args       args
		check      func(t *testing.T, resp checkoutSummaryResponse)
		checkTrace func(t *testing.T, tr traceResponse)
	}{
		{
			name: "returns checkout summary for valid user and product",
			args: args{userID: "u1", productID: "p1"},
			check: func(t *testing.T, resp checkoutSummaryResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "u1", resp.Summary.User.ID)
				require.Equal(t, "Alice", resp.Summary.User.Name)
				require.Equal(t, "gold", resp.Summary.User.Tier)
				require.True(t, resp.Summary.StockStatus.Available)
				require.InDelta(t, 1080.0, resp.Summary.DiscountedPrice, 0.01)
			},
			checkTrace: func(t *testing.T, tr traceResponse) {
				require.Equal(t, "ok", tr.Status)
				require.Equal(t, "checkout", tr.NamespaceID)
				require.Equal(t, "checkout_summary", tr.GraphID)
				require.Len(t, tr.NodeTraces, 7)

				input := findNodeTrace(t, tr, "input")
				require.Equal(t, 0, input.StatusCode)
				require.Empty(t, input.Error)

				for _, name := range []string{
					"parse-request", "validate-user", "fetch-product",
					"check-inventory", "apply-discount", "build-summary",
				} {
					n := findNodeTrace(t, tr, name)
					require.Equal(t, http.StatusOK, n.StatusCode, "node %s", name)
					require.Empty(t, n.Error, "node %s", name)
				}
			},
		},
		{
			name: "returns correct discounted price for cheaper product",
			args: args{userID: "u2", productID: "p2"},
			check: func(t *testing.T, resp checkoutSummaryResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "Bob", resp.Summary.User.Name)
				require.InDelta(t, 22.5, resp.Summary.DiscountedPrice, 0.01)
			},
		},
		{
			// Graph fails midway: validate-user (level 2) returns 403 and stops execution.
			// Both validate-user and fetch-product run in parallel at level 2 before the
			// error is detected — so fetch-product completes successfully while validate-user fails.
			name: "returns 403 for blocked user — graph fails midway with partial level execution",
			args: args{userID: "blocked", productID: "p1"},
			check: func(t *testing.T, resp checkoutSummaryResponse) {
				require.Equal(t, http.StatusForbidden, resp.StatusCode)
			},
			checkTrace: func(t *testing.T, tr traceResponse) {
				require.Equal(t, "error", tr.Status)

				// Levels 0 and 1 fully executed.
				input := findNodeTrace(t, tr, "input")
				require.Equal(t, 0, input.StatusCode)
				require.Empty(t, input.Error)

				parseRequest := findNodeTrace(t, tr, "parse-request")
				require.Equal(t, http.StatusOK, parseRequest.StatusCode)
				require.Empty(t, parseRequest.Error)

				// Level 2: both nodes ran in parallel.
				// validate-user failed with 403.
				validateUser := findNodeTrace(t, tr, "validate-user")
				require.Equal(t, http.StatusForbidden, validateUser.StatusCode)
				require.NotEmpty(t, validateUser.Error)
			},
		},
		{
			name: "returns 400 when required params are missing",
			args: args{userID: "", productID: ""},
			check: func(t *testing.T, resp checkoutSummaryResponse) {
				require.Equal(t, http.StatusBadRequest, resp.StatusCode)
			},
		},
		{
			name: "returns 404 for unknown product",
			args: args{userID: "u1", productID: "p999"},
			check: func(t *testing.T, resp checkoutSummaryResponse) {
				require.Equal(t, http.StatusNotFound, resp.StatusCode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := getCheckoutSummary(t, tc.args.userID, tc.args.productID)
			tc.check(t, resp)
			if tc.checkTrace != nil {
				tr := getTrace(t, shortcutURL, resp.RequestID)
				tc.checkTrace(t, tr)
			}
		})
	}
}

type checkoutUser struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type checkoutStockStatus struct {
	ProductID string `json:"product_id"`
	Available bool   `json:"available"`
	Quantity  int    `json:"quantity"`
}

type checkoutSummaryBody struct {
	User            checkoutUser        `json:"user"`
	StockStatus     checkoutStockStatus `json:"stock_status"`
	DiscountedPrice float64             `json:"discounted_price"`
}

type checkoutSummaryResponse struct {
	StatusCode int
	RequestID  string
	Summary    checkoutSummaryBody
}

func getCheckoutSummary(t *testing.T, userID, productID string) checkoutSummaryResponse {
	t.Helper()

	url := shortcutURL + "/run/checkout/checkout/summary"
	sep := "?"
	if userID != "" {
		url += sep + "user_id=" + userID
		sep = "&"
	}
	if productID != "" {
		url += sep + "product_id=" + productID
	}

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	result := checkoutSummaryResponse{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-Id"),
	}

	if resp.StatusCode == http.StatusOK {
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.Summary))
	}

	return result
}
