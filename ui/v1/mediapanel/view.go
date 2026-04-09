package mediapanel

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
)

func (m *Model) View() string {
	panelShell := m.styles.panel.Width(m.width).Height(m.height).Render("")
	panelNav := m.renderPanelNav()
	m.activePanel().SetSize(m.width, m.height)
	listView := m.activePanel().View()
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(panelShell).ID("panel"),
		lipgloss.NewLayer(panelNav).X(m.width/2 - lipgloss.Width(panelNav)/2).Y(0).ID("panelNav"),
		lipgloss.NewLayer(listView).X(1).Y(lipgloss.Height(panelNav)).ID("list"),
	}
	return lipgloss.NewCompositor(layers...).Render()
}

func (m *Model) renderPanelNav() string {
	segments := []struct {
		label string
		kind  common.ListKind
	}{
		{label: "PL", kind: common.Playlists},
		{label: "TR", kind: common.Tracks},
		{label: "AL", kind: common.Albums},
		{label: "AR", kind: common.Artists},
	}

	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if m.activePanel().kind == segment.kind {
			parts = append(parts, m.styles.panelNavActive.Render(segment.label))
			continue
		}
		parts = append(parts, m.styles.panelNavMuted.Render(segment.label))
	}
	return m.styles.panelNav.Render(strings.Join(parts, " - "))
}

func (p *panel) View() string {
	p.activeList().SetSize(p.width, p.height)
	return p.activeList().View()
}

func (p *panel) SetSize(width, height int) {
	p.width = width
	p.height = height
}
