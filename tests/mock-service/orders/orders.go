package orders

import (
	"encoding/json"
	"net/http"
	"slices"
	"strconv"

	xslices "github.com/Gadzet005/shortcut/pkg/containers/slices"
	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

type GetTopOrdersResponse struct {
	Orders []UserOrder `json:"orders"`
}

type GetUsersRequest struct {
	IDs []string `json:"ids"`
}

type GetUsersResponse struct {
	Users []User `json:"users"`
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Order struct {
	ID     string `json:"id"`
	UserID string `json:"user_id"`
	Amount int    `json:"amount"`
}

type UserOrder struct {
	UserName string `json:"user_name"`
	UserID   string `json:"user_id"`
	OrderID  string `json:"order_id"`
	Amount   int    `json:"amount"`
}

func GetUsersByIDs(ctx *shortcut.Context) error {
	users := []User{
		{ID: "1", Name: "John Doe"},
		{ID: "2", Name: "Jane Smith"},
		{ID: "3", Name: "Jim Beam"},
		{ID: "4", Name: "Bob Brown"},
		{ID: "5", Name: "Alice Cooper"},
		{ID: "6", Name: "Charlie Chaplin"},
		{ID: "7", Name: "David Bowie"},
		{ID: "8", Name: "Eva Green"},
		{ID: "9", Name: "Frank Sinatra"},
		{ID: "10", Name: "George Clooney"},
		{ID: "11", Name: "Hannah Montana"},
		{ID: "12", Name: "Isaac Newton"},
	}

	var ids []string
	if err := ctx.GetJSONItem("ids", &ids); err != nil {
		return err
	}

	usersMap := make(map[string]User, len(users))
	for _, user := range users {
		usersMap[user.ID] = user
	}

	responseUsers := make([]User, len(ids))
	for i, id := range ids {
		responseUsers[i] = usersMap[id]
	}

	return shortcut.NewResponse().
		AddJSONItem("users", responseUsers).
		Send(ctx)
}

func GetTopOrders(ctx *shortcut.Context) error {
	orders := []Order{
		{ID: "1", UserID: "1", Amount: 100},
		{ID: "1", UserID: "1", Amount: 300},
		{ID: "1", UserID: "1", Amount: 300},
		{ID: "2", UserID: "2", Amount: 200},
		{ID: "2", UserID: "2", Amount: 50},
		{ID: "3", UserID: "3", Amount: 300},
		{ID: "4", UserID: "4", Amount: 400},
		{ID: "4", UserID: "4", Amount: 100},
		{ID: "5", UserID: "5", Amount: 500},
		{ID: "6", UserID: "11", Amount: 55},
		{ID: "6", UserID: "11", Amount: 555},
	}

	var httpRequest shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &httpRequest); err != nil {
		return err
	}

	limitRaw := httpRequest.Query.Get("limit")
	limit, err := strconv.Atoi(limitRaw)
	if err != nil {
		return shortcut.NewErrorWithCause(400, "failed to parse limit", err)
	}
	if limit < 0 {
		return shortcut.NewError(400, "limit must be positive")
	}

	slices.SortFunc(orders, func(a, b Order) int {
		return b.Amount - a.Amount
	})

	responseOrders := orders
	if limit != 0 && len(responseOrders) > limit {
		responseOrders = responseOrders[:limit]
	}

	userIDs := xslices.Map(responseOrders, func(order Order) string {
		return order.UserID
	})

	return shortcut.NewResponse().
		AddJSONItem("orders", responseOrders).
		AddJSONItem("user_ids", userIDs).
		Send(ctx)
}

func MergeOrdersAndUsers(ctx *shortcut.Context) error {
	var orders []Order
	if err := ctx.GetJSONItem("orders", &orders); err != nil {
		return err
	}

	var users []User
	if err := ctx.GetJSONItem("users", &users); err != nil {
		return err
	}

	usersMap := make(map[string]User, len(users))
	for _, user := range users {
		usersMap[user.ID] = user
	}

	responseOrders := make([]UserOrder, len(orders))
	for i, order := range orders {
		responseOrders[i] = UserOrder{
			UserName: usersMap[order.UserID].Name,
			UserID:   order.UserID,
			OrderID:  order.ID,
			Amount:   order.Amount,
		}
	}

	bodyRaw, err := json.Marshal(responseOrders)
	if err != nil {
		return err
	}

	httpResponse := shortcutapi.HttpResponse{
		StatusCode: http.StatusOK,
		Body:       bodyRaw,
	}

	return shortcut.NewResponse().
		AddJSONItem("http_response", httpResponse).
		Send(ctx)
}
