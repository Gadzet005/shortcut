package main

import (
	"strconv"

	"github.com/Gadzet005/shortcut/shortcut/pkg/shortcut"
	"go.uber.org/zap"
)

func echoHandler(ctx *shortcut.Context) error {
	data, ok := ctx.GetItem("req")
	if !ok {
		return shortcut.ErrItemNotFound
	}

	ctx.Logger().Debug("Echo handler", zap.Int("data_size", len(data)))

	return shortcut.NewResponse().
		AddItem("resp", data).
		Send(ctx)
}

func sumHandler(ctx *shortcut.Context) error {
	aRaw, ok := ctx.GetItem("a")
	if !ok {
		return shortcut.ErrItemNotFound
	}

	bRaw, ok := ctx.GetItem("b")
	if !ok {
		return shortcut.ErrItemNotFound
	}

	a, err := strconv.Atoi(string(aRaw))
	if err != nil {
		return err
	}
	b, err := strconv.Atoi(string(bRaw))
	if err != nil {
		return err
	}

	return shortcut.NewResponse().
		AddItem("sum", []byte(strconv.Itoa(a+b))).
		Send(ctx)
}
