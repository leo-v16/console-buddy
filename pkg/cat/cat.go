package cat

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type Cat struct {
	Frames []string
	Index  int
}

func New() Cat {
	return Cat{
		Frames: []string{
			` /\_/\  `,
			`( o.o ) `,
			` > ^ <  `,
			` /\_/\  `,
			`( o.o ) `,
			` > ^ <  `,
		},
	}
}

func (c *Cat) NextFrame() {
	c.Index = (c.Index + 1) % len(c.Frames)
}

func (c Cat) View() string {
	return c.Frames[c.Index]
}

type Msg struct{}

func Animate() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return Msg{}
	})
}
