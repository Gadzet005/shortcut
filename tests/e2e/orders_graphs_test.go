package e2e

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetTopOrders(t *testing.T) {
	type args struct {
		limit string
	}

	testCases := []struct {
		name  string
		args  args
		check func(t *testing.T, response getTopOrdersResponse)
	}{
		{
			name: "returns top 3 orders with user names",
			args: args{
				limit: "3",
			},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusOK, response.StatusCode)
				require.Len(t, response.Orders, 3)
				require.Equal(t, userOrder{UserName: "Hannah Montana", UserID: "11", OrderID: "6", Amount: 555}, response.Orders[0])
				require.Equal(t, userOrder{UserName: "Alice Cooper", UserID: "5", OrderID: "5", Amount: 500}, response.Orders[1])
				require.Equal(t, userOrder{UserName: "Bob Brown", UserID: "4", OrderID: "4", Amount: 400}, response.Orders[2])
			},
		},
		{
			name: "returns all orders when limit exceeds total",
			args: args{
				limit: "100",
			},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusOK, response.StatusCode)
				require.Len(t, response.Orders, 11)
			},
		},
		{
			name: "returns error when limit is not a number",
			args: args{
				limit: "not-a-number",
			},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusBadRequest, response.StatusCode)
			},
		},
		{
			name: "returns error when limit is negative",
			args: args{
				limit: "-1",
			},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusBadRequest, response.StatusCode)
			},
		},
		{
			name: "returns all orders when limit is not provided",
			args: args{
				limit: "",
			},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusOK, response.StatusCode)
				require.Len(t, response.Orders, 11)
			},
		},
		{
			name: "returns error when limit is 0",
			args: args{
				limit: "0",
			},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusBadRequest, response.StatusCode)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := getTopOrders(t, testCase.args.limit)
			testCase.check(t, response)
		})
	}
}

type getTopOrdersResponse struct {
	StatusCode int         `json:"status_code"`
	Orders     []userOrder `json:"orders"`
}

type userOrder struct {
	UserName string `json:"user_name"`
	UserID   string `json:"user_id"`
	OrderID  string `json:"order_id"`
	Amount   int    `json:"amount"`
}

func getTopOrders(t *testing.T, limit string) getTopOrdersResponse {
	t.Helper()

	resp, err := http.Get(shortcutURL + "/run/orders/orders/get-top-orders?limit=" + limit)
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	response := getTopOrdersResponse{
		StatusCode: resp.StatusCode,
	}

	if resp.StatusCode == http.StatusOK {
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&response.Orders))
	} else {
		errorResponse := make(map[string]string)
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&errorResponse))
		require.NotEmpty(t, errorResponse["error"], "error response should contain error field")
	}

	return response
}
