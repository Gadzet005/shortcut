package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Multi-level DAG:
// input ─► parse-request ─► validate-user ──────────────────────────────┐
//                       └─► fetch-product ─► check-inventory ────────────┼─► build-summary
//                                        └─► apply-discount ─────────────┘

func TestCheckoutSummary(t *testing.T) {
	type args struct {
		userID    string
		productID string
	}

	testCases := []struct {
		name  string
		args  args
		check func(t *testing.T, resp checkoutSummaryResponse)
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
			name: "returns 403 for blocked user",
			args: args{userID: "blocked", productID: "p1"},
			check: func(t *testing.T, resp checkoutSummaryResponse) {
				require.Equal(t, http.StatusForbidden, resp.StatusCode)
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

	result := checkoutSummaryResponse{StatusCode: resp.StatusCode}

	if resp.StatusCode == http.StatusOK {
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.Summary))
	}

	return result
}
