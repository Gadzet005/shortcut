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
		name       string
		args       args
		check      func(t *testing.T, response getTopOrdersResponse)
		checkTrace func(t *testing.T, tr traceResponse)
	}{
		{
			name: "returns top 3 orders with user names",
			args: args{limit: "3"},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusOK, response.StatusCode)
				require.Len(t, response.Orders, 3)
				require.Equal(t, userOrder{UserName: "Hannah Montana", UserID: "11", OrderID: "6", Amount: 555}, response.Orders[0])
				require.Equal(t, userOrder{UserName: "Alice Cooper", UserID: "5", OrderID: "5", Amount: 500}, response.Orders[1])
				require.Equal(t, userOrder{UserName: "Bob Brown", UserID: "4", OrderID: "4", Amount: 400}, response.Orders[2])
			},
			checkTrace: func(t *testing.T, tr traceResponse) {
				require.Equal(t, "ok", tr.Status)
				require.Equal(t, "orders", tr.NamespaceID)
				require.Equal(t, "get_top_orders", tr.GraphID)
				require.Len(t, tr.NodeTraces, 4)

				input := findNodeTrace(t, tr, "input")
				require.Equal(t, 0, input.StatusCode)
				require.Empty(t, input.Error)

				for _, name := range []string{"get-top-orders", "get-users-by-ids", "merge-orders-and-users"} {
					n := findNodeTrace(t, tr, name)
					require.Equal(t, http.StatusOK, n.StatusCode, "node %s", name)
					require.Empty(t, n.Error, "node %s", name)
				}
			},
		},
		{
			name: "returns all orders when limit exceeds total",
			args: args{limit: "100"},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusOK, response.StatusCode)
				require.Len(t, response.Orders, 11)
			},
		},
		{
			name: "returns error when limit is not a number",
			args: args{limit: "not-a-number"},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusBadRequest, response.StatusCode)
			},
		},
		{
			name: "returns error when limit is negative",
			args: args{limit: "-1"},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusBadRequest, response.StatusCode)
			},
		},
		{
			name: "returns all orders when limit is not provided",
			args: args{limit: ""},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusOK, response.StatusCode)
				require.Len(t, response.Orders, 11)
			},
		},
		{
			name: "returns error when limit is 0",
			args: args{limit: "0"},
			check: func(t *testing.T, response getTopOrdersResponse) {
				require.Equal(t, http.StatusBadRequest, response.StatusCode)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			response := getTopOrders(t, testCase.args.limit)
			testCase.check(t, response)
			if testCase.checkTrace != nil {
				tr := getTrace(t, shortcutURL, response.RequestID)
				testCase.checkTrace(t, tr)
			}
		})
	}
}

type getTopOrdersResponse struct {
	StatusCode int `json:"status_code"`
	RequestID  string
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
		RequestID:  resp.Header.Get("X-Request-Id"),
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
