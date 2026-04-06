package v1

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func (m *MediaCenter) View() string {
	cassette := m.cassettePlayer.View()
	cassetteH := lipgloss.Height(cassette)
	listW := 30
	listH := cassetteH
	m.visibleList.SetSize(listW, listH)
	mediaList := m.visibleList.View()
	v := lipgloss.JoinHorizontal(lipgloss.Top, mediaList, cassette)
	return v
}

func outerShell(width int, height int) string {
	lines := make([]string, 0, height+2)
	lines = append(lines, strings.Repeat(" ", width))
	for range height {
		fill := strings.Repeat(" ", width)
		lines = append(lines,fill)
	}
	lines = append(lines, strings.Repeat(" ", width))
	return  lipgloss.JoinVertical(lipgloss.Left, lines...)
}

