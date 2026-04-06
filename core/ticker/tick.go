package ticker

import (
	"time"

	tea "charm.land/bubbletea/v2"
)

type TickFastMsg struct {}
type TickSlowMsg struct {}
type TickMsg struct {}

func DoTickFast() tea.Cmd {
	return tea.Tick(180*time.Millisecond, func(t time.Time) tea.Msg {
		return TickFastMsg{}
	})
}

func DoTickSlow() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return TickSlowMsg{}
	})
}

func DoTick() tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func StartTicker() tea.Cmd {
	return tea.Batch(DoTickFast(), DoTickSlow(),DoTick())
}
