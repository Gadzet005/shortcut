package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

type WeatherData struct {
	City        string `json:"city"`
	Temperature int    `json:"temperature"`
	Condition   string `json:"condition"`
}

type TrafficData struct {
	City          string `json:"city"`
	CongestionPct int    `json:"congestion_pct"`
	Incidents     int    `json:"incidents"`
}

type EventData struct {
	City   string   `json:"city"`
	Events []string `json:"events"`
}

type DashboardReport struct {
	Weather WeatherData `json:"weather"`
	Traffic TrafficData `json:"traffic"`
	Events  EventData   `json:"events"`
}

func FetchWeather(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	city := request.Query.Get("city")
	if city == "" {
		city = "Moscow"
	}

	weather := WeatherData{City: city, Temperature: 18, Condition: "Sunny"}
	return shortcut.NewResponse().
		AddJSONItem("weather", weather).
		Send(ctx)
}

func FetchTraffic(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	city := request.Query.Get("city")
	if city == "" {
		city = "Moscow"
	}

	traffic := TrafficData{City: city, CongestionPct: 42, Incidents: 3}
	return shortcut.NewResponse().
		AddJSONItem("traffic", traffic).
		Send(ctx)
}

func FetchEvents(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	city := request.Query.Get("city")
	if city == "" {
		city = "Moscow"
	}

	if city == "error" {
		return shortcut.NewError(http.StatusInternalServerError, "events service unavailable")
	}

	events := EventData{City: city, Events: []string{"Concert at Luzhniki", "Tech Conference"}}
	return shortcut.NewResponse().
		AddJSONItem("events", events).
		Send(ctx)
}

func AggregateReport(ctx *shortcut.Context) error {
	var weather WeatherData
	if err := ctx.GetJSONItem("weather", &weather); err != nil {
		return err
	}

	var traffic TrafficData
	if err := ctx.GetJSONItem("traffic", &traffic); err != nil {
		return err
	}

	var events EventData
	if err := ctx.GetJSONItem("events", &events); err != nil {
		return err
	}

	report := DashboardReport{
		Weather: weather,
		Traffic: traffic,
		Events:  events,
	}

	bodyRaw, err := json.Marshal(report)
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
