package main

import (
	"github.com/imafish/ttf"
)

type quitCommand struct{}

func (c *quitCommand) Handle(app *ttf.Application) error {
	app.Printf("Quit? press 'y' to confirm, any other key to cancel.")
	userInput, err := app.ReadKey()
	if err != nil {
		return err
	}
	if userInput == "y" {
		app.Printf("Quitting application...")
		app.Quit()
	} else {
		app.Printf("Quit cancelled.")
	}
	return nil
}

func (c *quitCommand) Description() string {
	return "Quit the application. Press 'y' to confirm."
}

func (c *quitCommand) Title() string {
	return "Quit the application"
}

func (c *quitCommand) Type() ttf.CommandType {
	return ttf.CommandTypeOneGo
}

func (c *quitCommand) Trigger() string {
	return "q"
}

func (c *quitCommand) Refresh() {
	// No refresh needed for quit command
}
