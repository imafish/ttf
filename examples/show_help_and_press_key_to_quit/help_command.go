package main

import (
	"fmt"

	"github.com/gizak/termui/v3/widgets"
	"github.com/imafish/ttf"
	"golang.org/x/exp/slog"
)

type helpCommandState int

const (
	helpCommandStateUninitialized helpCommandState = iota
	helpCommandStateVisible
	helpCommandStateHidden
)
const helpText = `Help Information:
This application demonstrates how to show help information and toggle its visibility using a key press.
Press 'h' to toggle the help information.
Press 'q' to quit the application.`

type helpCommand struct {
	state  helpCommandState
	widget *ttf.Widget // The widget that displays the help information
}

func (c *helpCommand) Handle(app *ttf.Application) error {
	if c.state == helpCommandStateUninitialized {
		drawable := widgets.NewParagraph()
		drawable.Title = c.Title()
		drawable.Text = helpText
		widget, err := app.NewAndAddWidget(c, drawable, 4)
		c.widget = widget
		if err != nil {
			slog.Error("Failed to create help widget", "error", err)
			return err
		}
		c.state = helpCommandStateHidden
	}

	var err error
	switch c.state {
	case helpCommandStateHidden:
		c.state = helpCommandStateVisible
		err = c.showWidget(app)
	case helpCommandStateVisible:
		c.state = helpCommandStateHidden
		err = c.hideWidget(app)
	default:
		slog.Error("Invalid help command state", "state", c.state)
		return fmt.Errorf("invalid help command state: %v", c.state)
	}

	return err
}

func (c *helpCommand) showWidget(app *ttf.Application) error {
	err := app.ShowWidget(c.widget)
	return err
}
func (c *helpCommand) hideWidget(app *ttf.Application) error {
	err := app.HideWidget(c.widget)
	return err
}

func (c *helpCommand) Description() string {
	return "Show help information. Press 'h' to toggle visibility."
}

func (c *helpCommand) Title() string {
	return "Help"
}

func (c *helpCommand) Type() ttf.CommandType {
	return ttf.CommandTypeStreaming
}

func (c *helpCommand) Trigger() string {
	return "h"
}

func (c *helpCommand) Refresh() {
	if c.state == helpCommandStateVisible {
		slog.Debug("Refreshing help widget", "title", c.Title())
		c.widget.Drawable.(*widgets.Paragraph).Text = helpText
		c.widget.Drawable.(*widgets.Paragraph).Title = c.Title()
		// Redraw the widget to reflect the updated text
		c.widget.Render()
	} else {
		slog.Debug("Help widget is not visible, skipping refresh", "title", c.Title())
	}
}
