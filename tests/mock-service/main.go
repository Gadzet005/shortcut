package main

import (
	"github.com/Gadzet005/shortcut/pkg/app/lifecycle"
)

func main() {
	lifecycle.Run(newService())
}
