package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

// Diamond graph: input → fetch-product → fetch-inventory ─┐
//                                      └─► fetch-pricing  ─┴─► build-detail

func TestGetProductDetail(t *testing.T) {
	type args struct {
		productID string
	}

	testCases := []struct {
		name  string
		args  args
		check func(t *testing.T, resp getProductDetailResponse)
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
			name: "returns 404 for unknown product",
			args: args{productID: "999"},
			check: func(t *testing.T, resp getProductDetailResponse) {
				require.Equal(t, http.StatusNotFound, resp.StatusCode)
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

	result := getProductDetailResponse{StatusCode: resp.StatusCode}

	if resp.StatusCode == http.StatusOK {
		require.Contains(t, resp.Header.Get("Content-Type"), "application/json")
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&result.Detail))
	}

	return result
}
