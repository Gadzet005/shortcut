package handlers

import (
	"context"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

func GetCart(ctx *shortcut.Context) error {
	var req shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &req); err != nil {
		return err
	}

	userID := req.Query.Get("user_id")
	if userID == "" {
		return shortcut.NewError(400, "user_id is required")
	}

	rows, err := pool.Query(context.Background(),
		`SELECT product_id FROM demo_cart_items
		 WHERE user_id = $1 AND qty > 0
		 ORDER BY product_id`, userID)
	if err != nil {
		return shortcut.NewErrorWithCause(500, "query cart", err)
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var pid string
		if err := rows.Scan(&pid); err != nil {
			return shortcut.NewErrorWithCause(500, "scan cart", err)
		}
		ids = append(ids, pid)
	}
	if err := rows.Err(); err != nil {
		return shortcut.NewErrorWithCause(500, "iterate cart", err)
	}

	return shortcut.NewResponse().
		AddJSONItem("cart_product_ids", ids).
		Send(ctx)
}
