package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/gin-gonic/gin"
)

func RemoveCartItem(ctx *shortcut.Context) error {
	var user User
	if err := ctx.GetJSONItem("user", &user); err != nil {
		return err
	}
	var product Product
	if err := ctx.GetJSONItem("product", &product); err != nil {
		return err
	}

	if _, err := pool.Exec(context.Background(),
		`UPDATE demo_cart_items SET qty = GREATEST(qty - 1, 0)
		 WHERE user_id = $1 AND product_id = $2`,
		user.ID, product.ID); err != nil {
		return shortcut.NewErrorWithCause(500, "remove cart item", err)
	}

	body, _ := json.Marshal(map[string]string{
		"status":     "removed",
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

func RevertRemoveCartItem(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad form"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "reverted"})
}
