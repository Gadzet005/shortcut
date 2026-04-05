package store

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

type Item struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CreatedAt string  `json:"created_at"`
}

var (
	mu     sync.Mutex
	items  []Item
	nextID = 1
)

// ValidateItem reads name and price from the original HTTP request query params,
// validates them, and passes them as separate items to the next node.
func ValidateItem(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	name := request.Query.Get("name")
	if name == "" {
		return shortcut.NewError(http.StatusBadRequest, "name is required")
	}

	priceStr := request.Query.Get("price")
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil || price <= 0 {
		return shortcut.NewError(http.StatusBadRequest, "price must be a positive number")
	}

	return shortcut.NewResponse().
		AddJSONItem("name", name).
		AddJSONItem("price", price).
		Send(ctx)
}

// SaveItem writes the validated item to in-memory storage and returns it.
func SaveItem(ctx *shortcut.Context) error {
	var name string
	if err := ctx.GetJSONItem("name", &name); err != nil {
		return err
	}

	var price float64
	if err := ctx.GetJSONItem("price", &price); err != nil {
		return err
	}

	mu.Lock()
	item := Item{
		ID:        nextID,
		Name:      name,
		Price:     price,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	items = append(items, item)
	nextID++
	mu.Unlock()

	bodyRaw, err := json.Marshal(item)
	if err != nil {
		return err
	}

	httpResponse := shortcutapi.HttpResponse{
		StatusCode: http.StatusCreated,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       bodyRaw,
	}

	return shortcut.NewResponse().
		AddJSONItem("http_response", httpResponse).
		Send(ctx)
}

// GetAllItems returns all items currently in memory.
func GetAllItems(ctx *shortcut.Context) error {
	mu.Lock()
	snapshot := make([]Item, len(items))
	copy(snapshot, items)
	mu.Unlock()

	bodyRaw, err := json.Marshal(snapshot)
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

// DeleteAllItems removes all items from memory and returns the count of deleted items.
func DeleteAllItems(ctx *shortcut.Context) error {
	mu.Lock()
	count := len(items)
	items = nil
	nextID = 1
	mu.Unlock()

	type result struct {
		DeletedCount int `json:"deleted_count"`
	}

	bodyRaw, err := json.Marshal(result{DeletedCount: count})
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
