package mediacenter

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	playerView := m.player.View()
	playerW, playerH := lipgloss.Size(playerView)
	var mediaList string
	listW := 0
	if m.mediaListOpen {
		listW = 30
		m.mediaPanel.SetSize(listW, playerH)
		mediaList = m.mediaPanel.View()
	}
	m.displayScreen.SetSize(listW+playerW, 3)
	content := lipgloss.JoinHorizontal(lipgloss.Top, mediaList, playerView)
	content = lipgloss.JoinVertical(lipgloss.Left, m.displayScreen.View(), content)
	return lipgloss.NewStyle().BorderStyle(lipgloss.HiddenBorder()).Render(content)
}
