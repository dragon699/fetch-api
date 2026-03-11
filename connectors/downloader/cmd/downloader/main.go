package main

import (
	"os"

	"connector-downloader/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		os.Exit(1)
	}
}
