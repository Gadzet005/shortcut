package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/gin-gonic/gin"
)

type productRef struct {
	ID string `json:"id"`
}

var jsonHeaders = map[string][]string{"Content-Type": {"application/json"}}

func cartMutation(ctx *shortcut.Context, status, errLabel, sql string) error {
	var user User
	if err := ctx.GetJSONItem("user", &user); err != nil {
		return err
	}
	var product productRef
	if err := ctx.GetJSONItem("product", &product); err != nil {
		return err
	}

	if _, err := pool.Exec(context.Background(), sql, user.ID, product.ID); err != nil {
		return shortcut.NewErrorWithCause(500, errLabel, err)
	}

	body, _ := json.Marshal(map[string]string{
		"status":     status,
		"user_id":    user.ID,
		"product_id": product.ID,
	})
	return shortcut.NewResponse().
		AddJSONItem("http_response", shortcutapi.HttpResponse{
			StatusCode: http.StatusOK,
			Headers:    jsonHeaders,
			Body:       body,
		}).
		Send(ctx)
}

func RevertCartItem(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad form"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "reverted"})
}
