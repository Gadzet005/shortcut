package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/gin-gonic/gin"
)

type Product struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

func AddCartItem(ctx *shortcut.Context) error {
	var user User
	if err := ctx.GetJSONItem("user", &user); err != nil {
		return err
	}
	var product Product
	if err := ctx.GetJSONItem("product", &product); err != nil {
		return err
	}

	if _, err := pool.Exec(context.Background(),
		`INSERT INTO demo_cart_items (user_id, product_id, qty) VALUES ($1, $2, 1)
		 ON CONFLICT (user_id, product_id) DO UPDATE
		   SET qty = demo_cart_items.qty + 1`,
		user.ID, product.ID); err != nil {
		return shortcut.NewErrorWithCause(500, "insert cart item", err)
	}

	body, _ := json.Marshal(map[string]string{
		"status":     "added",
		"user_id":    user.ID,
		"product_id": product.ID,
	})

	return shortcut.NewResponse().
		AddJSONItem("http_response", shortcutapi.HttpResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       body,
		}).
		Send(ctx)
}

func RevertAddCartItem(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad form"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "reverted"})
}
