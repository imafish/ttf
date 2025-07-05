package ttf

import ui "github.com/gizak/termui/v3"

type WidgetManager interface {
	NewWidget(command Command, widget ui.Drawable) (*Widget, error)
}
