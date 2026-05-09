package handlers

import (
	"context"
	"errors"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func GetUser(ctx *shortcut.Context) error {
	var req shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &req); err != nil {
		return err
	}

	userID := req.Query.Get("user_id")
	if userID == "" {
		return shortcut.NewError(400, "user_id is required")
	}

	var user User
	err := pool.QueryRow(context.Background(),
		`SELECT id, name, email FROM demo_users WHERE id = $1`, userID).
		Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return shortcut.NewError(404, "user not found")
		}
		return shortcut.NewErrorWithCause(500, "query user", err)
	}

	return shortcut.NewResponse().
		AddJSONItem("user", user).
		Send(ctx)
}
