package common

import tea "charm.land/bubbletea/v2"

type Component interface {
	Update(tea.Msg) tea.Cmd
	View() string
}

type Sizable interface {
	SetSize(width, height int)
}
