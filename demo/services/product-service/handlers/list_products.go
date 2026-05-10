package handlers

import (
	"context"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
)

func ListProducts(ctx *shortcut.Context) error {
	var ids []string
	if err := ctx.GetJSONItem("recommendations", &ids); err != nil {
		return err
	}

	products, err := selectProductsByIDs(context.Background(), ids)
	if err != nil {
		return shortcut.NewErrorWithCause(500, "query products", err)
	}

	return shortcut.NewResponse().
		AddJSONItem("products", products).
		Send(ctx)
}

func selectProductsByIDs(ctx context.Context, ids []string) ([]Product, error) {
	if len(ids) == 0 {
		return []Product{}, nil
	}

	rows, err := pool.Query(ctx,
		`SELECT id, name, description, price, image_url
		 FROM demo_products
		 WHERE id = ANY($1)`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byID := make(map[string]Product, len(ids))
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price, &p.ImageURL); err != nil {
			return nil, err
		}
		byID[p.ID] = p
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	products := make([]Product, 0, len(ids))
	for _, id := range ids {
		if p, ok := byID[id]; ok {
			products = append(products, p)
		}
	}
	return products, nil
}
