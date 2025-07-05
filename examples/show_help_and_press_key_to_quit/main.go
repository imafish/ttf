package main

import (
	"log/slog"
	"os"

	"github.com/imafish/ttf"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	slog.SetDefault(logger)

	app := ttf.NewApplication("Welcome to TTF Application")
	app.Initialize()
	app.AddCommand(&quitCommand{}, &helpCommand{})
	app.Run()
}
