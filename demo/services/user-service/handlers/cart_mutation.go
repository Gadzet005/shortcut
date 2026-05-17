package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
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

const (
	revertSelectFailureSQL = `SELECT request_body FROM failures WHERE request_id = $1`
	revertDeleteFailureSQL = `DELETE FROM failures WHERE request_id = $1`
)

func RevertAddCartItem(c *gin.Context) {
	revertByFailure(c, removeCartItemSQL, "add", "removed")
}

func RevertRemoveCartItem(c *gin.Context) {
	revertByFailure(c, addCartItemSQL, "remove", "restored")
}

func revertByFailure(c *gin.Context, compensatingSQL, originalAction, revertStatus string) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad form"})
		return
	}
	requestID := c.Request.PostFormValue("request_id")
	if requestID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "request_id is required"})
		return
	}

	ctx := c.Request.Context()
	var requestBody []byte
	err := pool.QueryRow(ctx, revertSelectFailureSQL, requestID).Scan(&requestBody)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "failure not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "load failure: " + err.Error()})
		return
	}

	var req shortcutapi.HttpRequest
	if err := json.Unmarshal(requestBody, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "decode failure body"})
		return
	}

	userID := req.Query.Get("user_id")
	productID := req.Query.Get("product_id")
	if userID == "" || productID == "" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "missing user_id or product_id"})
		return
	}

	if _, err := pool.Exec(ctx, compensatingSQL, userID, productID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "compensate " + originalAction + ": " + err.Error()})
		return
	}
	if _, err := pool.Exec(ctx, revertDeleteFailureSQL, requestID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failure: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     revertStatus,
		"request_id": requestID,
		"user_id":    userID,
		"product_id": productID,
	})
}
