package failurerec

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
	"github.com/gin-gonic/gin"
)

var (
	mu     sync.Mutex
	counts = make(map[string]int)
)

func incr(key string) int {
	mu.Lock()
	defer mu.Unlock()
	counts[key]++
	return counts[key]
}

func get(key string) int {
	mu.Lock()
	defer mu.Unlock()
	return counts[key]
}

// StepA succeeds and tracks how many times it ran for the given session.
func StepA(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}
	sessionID := request.Query.Get("session_id")
	incr(sessionID + "/step-a-runs")

	return shortcut.NewResponse().
		AddJSONItem("session_id", sessionID).
		Send(ctx)
}

// StepB always fails with 400 (4xx is not retried, so we get exactly one
// attempt before the failure strategy fires).
func StepB(ctx *shortcut.Context) error {
	var sessionID string
	_ = ctx.GetJSONItem("session_id", &sessionID)
	incr(sessionID + "/step-b-runs")
	return shortcut.NewError(http.StatusBadRequest, "step-b broken")
}

// RevertStepA is the compensation endpoint for StepA. The graph engine POSTs
// a form with the request_id field; we count one revert per request_id.
func RevertStepA(c *gin.Context) {
	if err := c.Request.ParseForm(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad form"})
		return
	}
	requestID := c.PostForm("request_id")
	incr(requestID + "/step-a-reverts")
	c.JSON(http.StatusOK, gin.H{"status": "reverted"})
}

// FlakyFinal fails the first `fail_count` calls and succeeds afterwards.
// Used to verify that the finish strategy synchronously re-runs failed nodes.
func FlakyFinal(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}
	sessionID := request.Query.Get("session_id")
	failCount, _ := strconv.Atoi(request.Query.Get("fail_count"))

	n := incr(sessionID + "/flaky-final-calls")
	if n <= failCount {
		// 4xx so the executor does not auto-retry inside one Run invocation;
		// the test exercises Finish strategy, not retry backoff.
		return shortcut.NewError(http.StatusBadRequest, "transient")
	}

	bodyRaw, err := json.Marshal(map[string]int{"total_attempts": n})
	if err != nil {
		return err
	}
	return shortcut.NewResponse().
		AddJSONItem("http_response", shortcutapi.HttpResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       bodyRaw,
		}).Send(ctx)
}

// GetStats reads the counter for ?key=... and returns it as JSON.
func GetStats(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}
	key := request.Query.Get("key")
	bodyRaw, err := json.Marshal(map[string]int{"count": get(key)})
	if err != nil {
		return err
	}
	return shortcut.NewResponse().
		AddJSONItem("http_response", shortcutapi.HttpResponse{
			StatusCode: http.StatusOK,
			Headers:    map[string][]string{"Content-Type": {"application/json"}},
			Body:       bodyRaw,
		}).Send(ctx)
}
