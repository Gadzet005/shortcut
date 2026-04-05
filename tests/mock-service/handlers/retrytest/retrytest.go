package retrytest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

var (
	countersMu   sync.Mutex
	callCounters = make(map[string]int)
)

// FlakyEndpoint simulates a service that fails a configurable number of times
// before succeeding. Query parameters (carried in the original http_request):
//
//   - session_id  — isolates counters between test cases (use t.Name() or similar)
//   - fail_count  — number of times to return an error before succeeding (default 0)
//   - fail_status — HTTP status code to return on failure (default 500)
//
// On success the response body is {"total_attempts": N}.
// This lets tests verify how many attempts were actually made.
func FlakyEndpoint(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	sessionID := request.Query.Get("session_id")
	failCount, _ := strconv.Atoi(request.Query.Get("fail_count"))
	failStatus := http.StatusInternalServerError
	if s := request.Query.Get("fail_status"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			failStatus = v
		}
	}

	countersMu.Lock()
	callCounters[sessionID]++
	count := callCounters[sessionID]
	countersMu.Unlock()

	if count <= failCount {
		return shortcut.NewError(failStatus, "transient failure")
	}

	type result struct {
		TotalAttempts int `json:"total_attempts"`
	}
	bodyRaw, err := json.Marshal(result{TotalAttempts: count})
	if err != nil {
		return err
	}

	httpResponse := shortcutapi.HttpResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       bodyRaw,
	}
	return shortcut.NewResponse().
		AddJSONItem("http_response", httpResponse).
		Send(ctx)
}
