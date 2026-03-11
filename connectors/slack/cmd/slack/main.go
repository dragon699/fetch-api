package main

import (
	"os"

	"connector-slack/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
