package ttf

type CommandType int

const (
	CommandTypeOneGo CommandType = iota
	CommandTypeStreaming
)

type Command interface {
	Handle(*Application) error
	Type() CommandType
	Title() string
	Trigger() string
	Description() string
	Refresh()
}
