package ttf

import (
	"fmt"
	"log/slog"

	ui "github.com/gizak/termui/v3"
)

type Widget struct {
	ui.Drawable
	Title string

	weight  int // 1~10, used to calculate the height of the widget in the grid
	visible bool
}

func NewWidget(drawable ui.Drawable, title string, weight int) (*Widget, error) {
	if drawable == nil {
		return nil, fmt.Errorf("drawable cannot be nil")
	}
	if weight < 1 || weight > 10 {
		return nil, fmt.Errorf("weight must be between 1 and 10")
	}

	return &Widget{
		Drawable: drawable,
		Title:    title,

		weight:  weight,
		visible: false,
	}, nil
}

func (w *Widget) Render() {
	if !w.visible {
		slog.Debug("Skipping rendering for hidden widget", "title", w.Title)
		return // Do not render if the widget is not visible
	}
	ui.Render(w.Drawable)
}
