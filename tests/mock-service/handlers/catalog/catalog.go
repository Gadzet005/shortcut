package catalog

import (
	"encoding/json"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

type Product struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

type Inventory struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
	Available bool   `json:"available"`
}

type Pricing struct {
	ProductID string  `json:"product_id"`
	Price     float64 `json:"price"`
	Currency  string  `json:"currency"`
}

type ProductDetail struct {
	Product   Product   `json:"product"`
	Inventory Inventory `json:"inventory"`
	Pricing   Pricing   `json:"pricing"`
}

var products = map[string]Product{
	"1": {ID: "1", Name: "Widget A", Category: "Widgets"},
	"2": {ID: "2", Name: "Gadget B", Category: "Gadgets"},
	"3": {ID: "3", Name: "Gizmo C", Category: "Gizmos"},
}

var inventories = map[string]Inventory{
	"1": {ProductID: "1", Quantity: 100, Available: true},
	"2": {ProductID: "2", Quantity: 0, Available: false},
	"3": {ProductID: "3", Quantity: 50, Available: true},
}

var pricings = map[string]Pricing{
	"1": {ProductID: "1", Price: 9.99, Currency: "USD"},
	"2": {ProductID: "2", Price: 19.99, Currency: "USD"},
	"3": {ProductID: "3", Price: 4.99, Currency: "USD"},
}

func FetchProduct(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	productID := request.Query.Get("product_id")
	product, ok := products[productID]
	if !ok {
		return shortcut.NewError(http.StatusNotFound, "product not found")
	}

	return shortcut.NewResponse().
		AddJSONItem("product", product).
		AddJSONItem("product_id", productID).
		Send(ctx)
}

func FetchInventory(ctx *shortcut.Context) error {
	var productID string
	if err := ctx.GetJSONItem("product_id", &productID); err != nil {
		return err
	}

	inventory, ok := inventories[productID]
	if !ok {
		return shortcut.NewError(http.StatusNotFound, "inventory not found")
	}

	return shortcut.NewResponse().
		AddJSONItem("inventory", inventory).
		Send(ctx)
}

func FetchPricing(ctx *shortcut.Context) error {
	var productID string
	if err := ctx.GetJSONItem("product_id", &productID); err != nil {
		return err
	}

	pricing, ok := pricings[productID]
	if !ok {
		return shortcut.NewError(http.StatusNotFound, "pricing not found")
	}

	return shortcut.NewResponse().
		AddJSONItem("pricing", pricing).
		Send(ctx)
}

func BuildDetail(ctx *shortcut.Context) error {
	var product Product
	if err := ctx.GetJSONItem("product", &product); err != nil {
		return err
	}

	var inventory Inventory
	if err := ctx.GetJSONItem("inventory", &inventory); err != nil {
		return err
	}

	var pricing Pricing
	if err := ctx.GetJSONItem("pricing", &pricing); err != nil {
		return err
	}

	detail := ProductDetail{
		Product:   product,
		Inventory: inventory,
		Pricing:   pricing,
	}

	bodyRaw, err := json.Marshal(detail)
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
