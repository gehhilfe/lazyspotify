package mediapanel

import (
	"charm.land/lipgloss/v2"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/medialist"
)

type styles struct {
	panel          lipgloss.Style
	panelNav       lipgloss.Style
	panelNavActive lipgloss.Style
	panelNavMuted  lipgloss.Style
}

type Model struct {
	panels []panel
	active int
	styles styles
	width  int
	height int
}

type panel struct {
	kind   common.ListKind
	lists  utils.Stack[medialist.Model]
	width  int
	height int
}

func NewModel() Model {
	kinds := []common.ListKind{
		common.Playlists,
		common.Tracks,
		common.Albums,
		common.Artists,
	}
	panels := common.MapSlice(kinds, newPanel)
	return Model{
		panels: panels,
		active: 0,
		styles: defaultStyles(),
	}
}

func defaultStyles() styles {
	return styles{
		panel:    lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()),
		panelNav: lipgloss.NewStyle(),
		panelNavActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true),
		panelNavMuted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
	}
}

func newPanel(kind common.ListKind) panel {
	p := panel{kind: kind}
	p.lists.Push(medialist.NewModel(kind))
	return p
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m *Model) activePanel() *panel {
	return &m.panels[m.active]
}
