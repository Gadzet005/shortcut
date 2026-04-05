package checkout

import (
	"encoding/json"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Tier string `json:"tier"`
}

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type StockStatus struct {
	ProductID string `json:"product_id"`
	Available bool   `json:"available"`
	Quantity  int    `json:"quantity"`
}

type CheckoutSummary struct {
	User            User        `json:"user"`
	StockStatus     StockStatus `json:"stock_status"`
	DiscountedPrice float64     `json:"discounted_price"`
}

var users = map[string]User{
	"u1": {ID: "u1", Name: "Alice", Tier: "gold"},
	"u2": {ID: "u2", Name: "Bob", Tier: "silver"},
}

var checkoutProducts = map[string]Product{
	"p1": {ID: "p1", Name: "Laptop", Price: 1200.00},
	"p2": {ID: "p2", Name: "Mouse", Price: 25.00},
}

func ParseRequest(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	userID := request.Query.Get("user_id")
	productID := request.Query.Get("product_id")
	if userID == "" || productID == "" {
		return shortcut.NewError(http.StatusBadRequest, "user_id and product_id are required")
	}

	return shortcut.NewResponse().
		AddJSONItem("user_id", userID).
		AddJSONItem("product_id", productID).
		Send(ctx)
}

func ValidateUser(ctx *shortcut.Context) error {
	var userID string
	if err := ctx.GetJSONItem("user_id", &userID); err != nil {
		return err
	}

	if userID == "blocked" {
		return shortcut.NewError(http.StatusForbidden, "user is blocked")
	}

	user, ok := users[userID]
	if !ok {
		return shortcut.NewError(http.StatusNotFound, "user not found")
	}

	return shortcut.NewResponse().
		AddJSONItem("user", user).
		Send(ctx)
}

func FetchProduct(ctx *shortcut.Context) error {
	var productID string
	if err := ctx.GetJSONItem("product_id", &productID); err != nil {
		return err
	}

	product, ok := checkoutProducts[productID]
	if !ok {
		return shortcut.NewError(http.StatusNotFound, "product not found")
	}

	return shortcut.NewResponse().
		AddJSONItem("product", product).
		Send(ctx)
}

func CheckInventory(ctx *shortcut.Context) error {
	var product Product
	if err := ctx.GetJSONItem("product", &product); err != nil {
		return err
	}

	status := StockStatus{
		ProductID: product.ID,
		Available: true,
		Quantity:  10,
	}

	return shortcut.NewResponse().
		AddJSONItem("stock_status", status).
		Send(ctx)
}

func ApplyDiscount(ctx *shortcut.Context) error {
	var product Product
	if err := ctx.GetJSONItem("product", &product); err != nil {
		return err
	}

	discountedPrice := product.Price * 0.9

	return shortcut.NewResponse().
		AddJSONItem("discounted_price", discountedPrice).
		Send(ctx)
}

func BuildSummary(ctx *shortcut.Context) error {
	var user User
	if err := ctx.GetJSONItem("user", &user); err != nil {
		return err
	}

	var stockStatus StockStatus
	if err := ctx.GetJSONItem("stock_status", &stockStatus); err != nil {
		return err
	}

	var discountedPrice float64
	if err := ctx.GetJSONItem("discounted_price", &discountedPrice); err != nil {
		return err
	}

	summary := CheckoutSummary{
		User:            user,
		StockStatus:     stockStatus,
		DiscountedPrice: discountedPrice,
	}

	bodyRaw, err := json.Marshal(summary)
	if err != nil {
		return err
	}

	httpResponse := shortcutapi.HttpResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       bodyRaw,
	}

	return shortcut.NewResponse().
		AddJSONItem("http_response", httpResponse).
		Send(ctx)
}
