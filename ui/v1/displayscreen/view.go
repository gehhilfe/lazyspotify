package displayscreen

import (
	"charm.land/lipgloss/v2"
)

func (m *Model) View() string {
	raw := m.display
	contentWidth := max(0, m.width-2)
	styled := m.styles.muted.Render(raw)
	if contentWidth > 0 {
		if lipgloss.Width(raw) > contentWidth {
			styled = m.styles.marquee.Render(m.scrollText(raw, contentWidth))
		}
		styled = lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center).Render(styled)
	}
	return m.styles.panel.Width(m.width).Height(m.height).Render(styled)
}
