package handlers

import (
	"math/rand/v2"
	"strconv"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

const defaultLimit = 3

func GetRecommendations(ctx *shortcut.Context) error {
	var req shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &req); err != nil {
		return err
	}

	excludeID := req.Query.Get("product_id")
	limit := defaultLimit
	if v := req.Query.Get("limit"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	pool := allIDs()
	if cat := categoryOf(excludeID); cat != "" {
		if filtered := idsByCategory(cat); len(filtered) > 0 {
			pool = filtered
		}
	}
	pool = removeID(pool, excludeID)

	return shortcut.NewResponse().
		AddJSONItem("recommendations", pickRandom(pool, limit)).
		Send(ctx)
}

func removeID(ids []string, id string) []string {
	if id == "" {
		return ids
	}
	out := make([]string, 0, len(ids))
	for _, v := range ids {
		if v != id {
			out = append(out, v)
		}
	}
	return out
}

func pickRandom(ids []string, n int) []string {
	if n > len(ids) {
		n = len(ids)
	}
	indices := rand.Perm(len(ids))[:n]
	out := make([]string, n)
	for i, idx := range indices {
		out[i] = ids[idx]
	}
	return out
}
