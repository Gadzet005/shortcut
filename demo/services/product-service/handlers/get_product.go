package handlers

import (
	"context"
	"errors"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/jackc/pgx/v5"
)

func GetProduct(ctx *shortcut.Context) error {
	var req shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &req); err != nil {
		return err
	}

	productID := req.Query.Get("product_id")
	if productID == "" {
		return shortcut.NewError(400, "product_id is required")
	}

	product, err := selectProduct(context.Background(), productID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return shortcut.NewError(404, "product not found")
		}
		return shortcut.NewErrorWithCause(500, "query product", err)
	}

	return shortcut.NewResponse().
		AddJSONItem("product", product).
		Send(ctx)
}

func selectProduct(ctx context.Context, id string) (Product, error) {
	var p Product
	err := pool.QueryRow(ctx,
		`SELECT id, name, description, price, image_url FROM demo_products WHERE id = $1`, id).
		Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL)
	return p, err
}
