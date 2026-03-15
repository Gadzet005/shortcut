package main

import (
	app "github.com/Gadzet005/shortcut/internal/app/shortcut"
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
)

func main() {
	lifecycle.Run(app.NewService())
}
