package medialist

import (
	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	listWidth := m.width - 4
	listHeight := m.height - 2
	footer := ""
	if m.pager.TotalPages > 1 {
		listHeight--
		footer = lipgloss.NewStyle().Width(max(0, listWidth)).Align(lipgloss.Center).Foreground(lipgloss.Color("8")).Render(m.pager.View())
	}
	if listHeight < 1 {
		listHeight = 1
	}
	m.list.SetSize(listWidth, listHeight)
	if m.state == Loading {
		m.list.Title = "Loading..."
		m.list.SetItems(nil)
	}
	if footer == "" {
		return m.list.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), footer)
}
