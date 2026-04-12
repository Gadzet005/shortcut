package counter

import (
	"encoding/json"
	"sync/atomic"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

var count atomic.Int64

func Get(ctx *shortcut.Context) error {
	n := count.Add(1)

	body, err := json.Marshal(map[string]int64{"count": n})
	if err != nil {
		return err
	}

	return shortcut.NewResponse().
		AddJSONItem("http_response", shortcutapi.HttpResponse{
			StatusCode: 200,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       body,
		}).
		Send(ctx)
}
