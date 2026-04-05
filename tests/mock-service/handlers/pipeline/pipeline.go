package pipeline

import (
	"encoding/json"

	"github.com/Gadzet005/shortcut/pkg/shortcut"
	shortcutapi "github.com/Gadzet005/shortcut/pkg/shortcut/api"
)

// Each step increments the passed value by 10.
// Step1 initialises value = 10 from the HTTP request.
// Final result after 5 steps: 50.

func Step1(ctx *shortcut.Context) error {
	var request shortcutapi.HttpRequest
	if err := ctx.GetJSONItem("request", &request); err != nil {
		return err
	}

	return shortcut.NewResponse().
		AddJSONItem("value", 10).
		Send(ctx)
}

func Step2(ctx *shortcut.Context) error {
	var value int
	if err := ctx.GetJSONItem("value", &value); err != nil {
		return err
	}

	return shortcut.NewResponse().
		AddJSONItem("value", value+10).
		Send(ctx)
}

func Step3(ctx *shortcut.Context) error {
	var value int
	if err := ctx.GetJSONItem("value", &value); err != nil {
		return err
	}

	return shortcut.NewResponse().
		AddJSONItem("value", value+10).
		Send(ctx)
}

func Step4(ctx *shortcut.Context) error {
	var value int
	if err := ctx.GetJSONItem("value", &value); err != nil {
		return err
	}

	return shortcut.NewResponse().
		AddJSONItem("value", value+10).
		Send(ctx)
}

func Step5(ctx *shortcut.Context) error {
	var value int
	if err := ctx.GetJSONItem("value", &value); err != nil {
		return err
	}

	finalValue := value + 10

	bodyRaw, err := json.Marshal(map[string]int{"result": finalValue})
	if err != nil {
		return err
	}

	httpResponse := shortcutapi.HttpResponse{
		StatusCode: 200,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       bodyRaw,
	}

	return shortcut.NewResponse().
		AddJSONItem("http_response", httpResponse).
		Send(ctx)
}
