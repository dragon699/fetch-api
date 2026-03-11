package main

import (
	"os"

	"connector-kube/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
