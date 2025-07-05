package ttf

import (
	"fmt"
	"image"
	"log"
	"log/slog"
	"strings"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type applicationState int

const (
	applicationStateNormal applicationState = iota //
	applicationStateTyping
	applicationStateSubcommand
)

// Application represents the main application state and holds commands.
// An application has a main output widget, a command input widget, and several on-the-fly widgets.
type Application struct {
	commands    map[string]Command
	state       applicationState
	quitFlag    bool // Flag to indicate if the application should quit
	mainGrid    *ui.Grid
	cmdWidget   *widgets.Paragraph
	cmdText     string
	mainWidget  *widgets.Paragraph
	mainText    []string
	widgets     []*Widget       // Widgets that are currently active in the application TODO: maybe this is not needed. Widgets can be managed in commands.
	uiEventChan <-chan ui.Event // Channel to receive UI events
	size        image.Point

	title string // Title of the application, can be used for display purposes
}

func NewApplication(title string) *Application {
	return &Application{
		commands: make(map[string]Command, 32),
		state:    applicationStateNormal,
		quitFlag: false,
		title:    title,
	}
}

func (app *Application) AddCommand(commands ...Command) error {
	for _, command := range commands {
		if command == nil {
			slog.Error("Attempted to add a nil command")
			return fmt.Errorf("cannot add nil command")
		}

		trigger := command.Trigger()
		if len(trigger) == 0 {
			slog.Error("Command trigger is empty", "command", command.Title())
			return fmt.Errorf("command trigger cannot be empty")
		}
		runeKey := rune(trigger[0])
		if _, exists := app.commands[string(runeKey)]; exists {
			slog.Error("Command trigger already exists", "trigger", trigger)
			return fmt.Errorf("command trigger '%s' already exists", trigger)
		}
		app.commands[string(runeKey)] = command
		slog.Info("Command added", "trigger", trigger, "title", command.Title())
	}
	return nil
}

func (app *Application) NewAndAddWidget(command Command, widget ui.Drawable, weight int) (*Widget, error) {
	if command == nil {
		return nil, fmt.Errorf("cannot create widget for nil command")
	}
	if widget == nil {
		return nil, fmt.Errorf("cannot create widget with nil drawable")
	}

	newWidget, err := NewWidget(widget, command.Title(), weight)
	if err != nil {
		slog.Error("Failed to create new widget", "command", command.Title(), "error", err)
		return nil, fmt.Errorf("failed to create widget: %w", err)
	}

	app.widgets = append(app.widgets, newWidget)
	slog.Info("New widget created", "command", command.Title(), "weight", weight)
	app.recalculateWidgetSizes()
	app.redrawWidgets()

	return newWidget, nil
}

func (app *Application) ShowWidget(widget *Widget) error {
	if widget == nil {
		return fmt.Errorf("cannot show nil widget")
	}
	if !widget.visible {
		widget.visible = true
		slog.Info("Showing widget", "command", widget.Title)
		app.recalculateWidgetSizes()
		app.redrawWidgets()
	} else {
		slog.Warn("Widget is already visible", "command", widget.Title)
	}
	return nil
}

func (app *Application) HideWidget(widget *Widget) error {
	if widget == nil {
		return fmt.Errorf("cannot hide nil widget")
	}
	if widget.visible {
		widget.visible = false
		slog.Info("Hiding widget", "command", widget.Title)
		app.recalculateWidgetSizes()
		app.redrawWidgets()
	} else {
		slog.Warn("Widget is already hidden", "command", widget.Title)
	}
	return nil
}

func (app *Application) Initialize() error {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}

	app.mainText = make([]string, 0, 100)
	app.widgets = make([]*Widget, 0, 10)

	app.size.X, app.size.Y = ui.TerminalDimensions()
	app.mainGrid = ui.NewGrid()

	app.mainWidget = widgets.NewParagraph()
	app.mainWidget.Title = app.title

	app.cmdWidget = widgets.NewParagraph()
	app.cmdWidget.Title = "Command"

	app.recalculateWidgetSizes()
	app.redrawWidgets()

	return nil
}

func (app *Application) redrawWidgets() {
	ui.Clear()

	ui.Render(app.mainGrid, app.cmdWidget) // Re-render the grids after resize

	// refresh main widget
	app.refreshMainWidget()

	for _, command := range app.commands {
		command.Refresh()
	}
}

func (app *Application) recalculateWidgetSizes() {
	const mainWidgetWeight = 6
	totalWeight := mainWidgetWeight
	visibleWidgets := make([]*Widget, 0, len(app.widgets))
	for _, widget := range app.widgets {
		if widget.visible {
			visibleWidgets = append(visibleWidgets, widget)
			totalWeight += widget.weight
		} else {
			slog.Debug("Widget is not visible, skipping", "command", widget.Title)
		}
	}

	calculatedMainWeight := float64(mainWidgetWeight) / float64(totalWeight)

	rows := make([]interface{}, 0, len(visibleWidgets)+1)
	rows = append(rows, ui.NewRow(calculatedMainWeight, app.mainWidget))

	for _, widget := range visibleWidgets {
		calculatedWeight := float64(widget.weight) / float64(totalWeight)
		rows = append(rows, ui.NewRow(calculatedWeight, widget.Drawable))
	}

	app.mainGrid.SetRect(0, 0, app.size.X, app.size.Y-3)
	app.mainGrid.Set(rows...)

	app.cmdWidget.SetRect(0, app.size.Y-3, app.size.X, app.size.Y)
}

func (app *Application) onTerminalResize(size image.Point) {
	app.size = size
	app.recalculateWidgetSizes()
	app.redrawWidgets()
}

func (app *Application) Run() error {
	app.uiEventChan = ui.PollEvents()
	for !app.quitFlag {
		event := app.handleDefaultEvent()
		if event != nil {
			keyboardInput := event.ID
			app.handleKeyboardInput(keyboardInput)
		}
	}
	return nil
}

func (app *Application) handleDefaultEvent() *ui.Event {
	e := <-app.uiEventChan
	switch e.Type {
	case ui.KeyboardEvent:
		if e.ID == "<C-c>" {
			app.Printf("Ctrl-C detected, quitting application...\n")
			app.Quit()
			return nil
		}

		return &e
	case ui.ResizeEvent:
		payload := e.Payload.(ui.Resize)
		app.onTerminalResize(image.Point{X: payload.Width, Y: payload.Height})
	case ui.MouseEvent:
		// NO handling for now
	}

	return nil
}

func (app *Application) handleKeyboardInput(id string) {
	switch app.state {
	case applicationStateNormal:
		app.keyboardInputNormal(id)

	case applicationStateSubcommand:
		app.keyboardInputTyping(id)
	case applicationStateTyping:
	}
}

func (app *Application) keyboardInputNormal(id string) {
	if id == ">" {
		app.state = applicationStateTyping
		app.cmdWidget.Text = "> "
		ui.Render(app.cmdWidget)
		return
	}

	command := app.findCommandByTrigger(id)
	if command == nil {
		app.Printf("key '%s' is not mapped to any command.", id)
		return
	}
	err := command.Handle(app)
	if err != nil {
		slog.Error("Error handling command", "command", command.Title(), "error", err)
		// TODO: print error in red
		app.Printf("Error handling command '%s': %v", command.Title(), err)
		return
	}
}

func (app *Application) keyboardInputTyping(id string) {
	panic("keyboardInputTyping not implemented yet")
}

func (app *Application) findCommandByTrigger(trigger string) Command {
	slog.Debug("Finding command by trigger", "trigger", trigger)
	if len(trigger) == 0 {
		return nil
	}
	command := app.commands[trigger]
	return command
}

func (app *Application) Quit() {
	ui.Close() // Close the termui library to clean up resources.
	app.quitFlag = true
}

func (app *Application) refreshMainWidget() {
	widgetSize := app.mainWidget.Size()

	displayText := ""
	if len(app.mainText) > 0 {
		startIndex := len(app.mainText) - widgetSize.Y + 2
		if startIndex >= len(app.mainText) {
			startIndex = len(app.mainText) - 1
		}
		if startIndex < 0 {
			startIndex = 0
		}
		displayText = strings.Join(app.mainText[startIndex:], "\n")
	}
	app.mainWidget.Text = displayText
	ui.Render(app.mainWidget) // Re-render the main grid and command widget
}

func (app *Application) Printf(format string, args ...interface{}) {
	stringToPrint := fmt.Sprintf(format, args...)
	stringsByLine := strings.Split(stringToPrint, "\n")
	app.mainText = append(app.mainText, stringsByLine...)

	app.refreshMainWidget()
	slog.Info("Printed to main widget", "text", stringToPrint)
}

func (app *Application) ReadKey() (string, error) {
	for {
		event := app.handleDefaultEvent()
		if event != nil {
			return event.ID, nil
		}
	}
}
