package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Diamond graph: input → fetch-product → fetch-inventory ─┐
//
//	└─► fetch-pricing  ─┴─► build-detail
func TestGetProductDetail(t *testing.T) {
	type args struct {
		productID string
	}

	testCases := []struct {
		name       string
		args       args
		check      func(t *testing.T, resp getProductDetailResponse)
		checkTrace func(t *testing.T, tr traceResponse)
	}{
		{
			name: "returns full product detail for valid product",
			args: args{productID: "1"},
			check: func(t *testing.T, resp getProductDetailResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "1", resp.Detail.Product.ID)
				require.Equal(t, "Widget A", resp.Detail.Product.Name)
				require.Equal(t, "Widgets", resp.Detail.Product.Category)
				require.Equal(t, 100, resp.Detail.Inventory.Quantity)
				require.True(t, resp.Detail.Inventory.Available)
				require.Equal(t, 9.99, resp.Detail.Pricing.Price)
				require.Equal(t, "USD", resp.Detail.Pricing.Currency)
			},
			checkTrace: func(t *testing.T, tr traceResponse) {
				require.Equal(t, "ok", tr.Status)
				require.Equal(t, "catalog", tr.NamespaceID)
				require.Equal(t, "get_product_detail", tr.GraphID)
				require.Len(t, tr.NodeTraces, 5)

				input := findNodeTrace(t, tr, "input")
				require.Equal(t, 0, input.StatusCode)
				require.Empty(t, input.Error)

				for _, name := range []string{"fetch-product", "fetch-inventory", "fetch-pricing", "build-detail"} {
					n := findNodeTrace(t, tr, name)
					require.Equal(t, http.StatusOK, n.StatusCode, "node %s", name)
					require.Empty(t, n.Error, "node %s", name)
				}
			},
		},
		{
			name: "returns correct detail for another product",
			args: args{productID: "2"},
			check: func(t *testing.T, resp getProductDetailResponse) {
				require.Equal(t, http.StatusOK, resp.StatusCode)
				require.Equal(t, "2", resp.Detail.Product.ID)
				require.Equal(t, "Gadget B", resp.Detail.Product.Name)
				require.False(t, resp.Detail.Inventory.Available)
				require.Equal(t, 19.99, resp.Detail.Pricing.Price)
			},
		},
		{
			// Graph fails at the first HTTP node (fetch-product returns 404).
			// Only input + fetch-product are traced — downstream nodes never execute.
			name: "returns 404 for unknown product — graph fails at first node",
			args: args{productID: "999"},
			check: func(t *testing.T, resp getProductDetailResponse) {
				require.Equal(t, http.StatusNotFound, resp.StatusCode)
			},
			checkTrace: func(t *testing.T, tr traceResponse) {
				require.Equal(t, "error", tr.Status)
				require.Len(t, tr.NodeTraces, 2) // only input + fetch-product ran

				input := findNodeTrace(t, tr, "input")
				require.Equal(t, 0, input.StatusCode)
				require.Empty(t, input.Error)

				fetchProduct := findNodeTrace(t, tr, "fetch-product")
				require.Equal(t, http.StatusNotFound, fetchProduct.StatusCode)
				require.NotEmpty(t, fetchProduct.Error)

				// These nodes must NOT be in the trace (never executed).
				for _, name := range []string{"fetch-inventory", "fetch-pricing", "build-detail"} {
					for _, nt := range tr.NodeTraces {
						require.NotEqual(t, name, nt.NodeID, "node %s should not have run", name)
					}
				}
			},
		},
		{
			name: "returns 404 when product_id is missing",
			args: args{productID: ""},
			check: func(t *testing.T, resp getProductDetailResponse) {
				require.Equal(t, http.StatusNotFound, resp.StatusCode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resp := getProductDetail(t, tc.args.productID)
			tc.check(t, resp)
			if tc.checkTrace != nil {
				tr := getTrace(t, shortcutURL, resp.RequestID)
				tc.checkTrace(t, tr)
			}
		})
	}
}

type catalogProduct struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type catalogInventory struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Available bool   `json:"available"`
}

type catalogPricing struct {
	ProductID string  `json:"product_id"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
}

type catalogProductDetail struct {
	Product   catalogProduct   `json:"product"`
	Inventory catalogInventory `json:"inventory"`
	Pricing   catalogPricing   `json:"pricing"`
}

type getProductDetailResponse struct {
	StatusCode int
	RequestID  string
	Detail     catalogProductDetail
}

func getProductDetail(t *testing.T, productID string) getProductDetailResponse {
	t.Helper()

	url := shortcutURL + "/run/catalog/catalog/get-product-detail"
	if productID != "" {
		url += "?product_id=" + productID
	}

	resp, err := http.Get(url)
	require.NoError(t, err)
	defer resp.Body.Close()

	result := getProductDetailResponse{
		StatusCode: resp.StatusCode,
		RequestID:  resp.Header.Get("X-Request-Id"),
	}

	if resp.StatusCode == http.StatusOK {
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.Detail))
	}

	return result
}
