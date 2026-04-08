package v1

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
)

type MediaPanel struct {
	panels []panel
	active int
	styles mediaPanelstyles
	width  int
	height int
}

type mediaPanelstyles struct {
	panel          lipgloss.Style
	panelNav       lipgloss.Style
	panelNavActive lipgloss.Style
	panelNavMuted  lipgloss.Style
}

func defaultMediaPanelStyles() mediaPanelstyles {
	return mediaPanelstyles{
		panel:    lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()),
		panelNav: lipgloss.NewStyle(),
		panelNavActive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("14")).
			Bold(true),
		panelNavMuted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")),
	}
}

type panel struct {
	kind   ListKind
	lists  utils.Stack[mediaList]
	width  int
	height int
}

func newPanel(kind ListKind) panel {
	p := panel{
		kind: kind,
	}
	p.lists.Push(newMediaList(p.kind))
	return p
}

func NewMediaPanel() MediaPanel {
	pl := []panel{
		newPanel(Playlists),
		newPanel(Tracks),
		newPanel(Albums),
		newPanel(Artists),
	}
	return MediaPanel{
		panels: pl,
		active: 0,
		styles: defaultMediaPanelStyles(),
	}
}

func (m *MediaPanel) SetSize(width int, height int) {
	m.width = width
	m.height = height
}

func (m *MediaPanel) GetActivePanel() *panel {
	return &m.panels[m.active]
}

func (m *MediaPanel) Update(msg tea.Msg) tea.Cmd {
	panel := m.GetActivePanel()
	cmd := panel.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, mediaCenterKeyMap.NextPanel):
			cmd = tea.Batch(cmd, m.activateNextPanel())
		}
	}
	return cmd
}

func (m *MediaPanel) View() string {
	panel := m.styles.panel.Width(m.width).Height(m.height).Render("")
	panelNav := m.renderPanelNav()
	m.GetActivePanel().SetSize(m.width, m.height)
	list := m.GetActivePanel().View()
	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(panel).ID("panel"),
		lipgloss.NewLayer(panelNav).X(m.width/2 - lipgloss.Width(panelNav)/2).Y(0).ID("panelNav"),
		lipgloss.NewLayer(list).X(1).Y(lipgloss.Height(panelNav)).ID("list"),
	}
	compositor := lipgloss.NewCompositor(layers...)
	panel = compositor.Render()
	return panel
}

func (m *MediaPanel) activateNextPanel() tea.Cmd {
	m.active = (m.active + 1) % len(m.panels)
	return m.GetActivePanel().Prepare()
}

func (m *MediaPanel) StartLoading() tea.Cmd {
	return m.GetActivePanel().GetActiveList().StartLoading()
}

func (m *MediaPanel) SetStatus(message string) tea.Cmd {
	return m.GetActivePanel().SetStatus(message)
}

func (m *MediaPanel) SetContent(entities []Entity, kind ListKind, pagination PaginationInfo, request MediaRequest) tea.Cmd {
	return m.GetActivePanel().SetContent(entities, kind, pagination, request)
}

func (m *MediaPanel) renderPanelNav() string {
	activePanel := m.GetActivePanel()
	segments := []struct {
		label string
		kind  ListKind
	}{
		{label: "PL", kind: Playlists},
		{label: "TR", kind: Tracks},
		{label: "AL", kind: Albums},
		{label: "AR", kind: Artists},
	}

	parts := make([]string, 0, len(segments))
	for _, segment := range segments {
		if activePanel.kind == segment.kind {
			parts = append(parts, m.styles.panelNavActive.Render(segment.label))
			continue
		}
		parts = append(parts, m.styles.panelNavMuted.Render(segment.label))
	}

	return m.styles.panelNav.Render(strings.Join(parts, " - "))
}

func (p *panel) GetActiveList() *mediaList {
	return p.lists.Peek()
}

func (p *panel) Update(msg tea.Msg) tea.Cmd {
	cmd := []tea.Cmd{
		p.GetActiveList().Update(msg),
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, mediaCenterKeyMap.Select):
			if item, ok := p.GetActiveList().list.SelectedItem().(mediaListItem); ok {
				cmd = append(cmd, item.entity.Action(p))
			}
		case key.Matches(msg, mediaCenterKeyMap.NextPage):
			if req, ok := p.GetActiveList().NextPageRequest(); ok {
				cmd = append(cmd, func() tea.Msg { return req })
			} else {
				cmd = append(cmd, p.GetActiveList().SetStatus("No next page"))
			}
		case key.Matches(msg, mediaCenterKeyMap.PrevPage):
			if req, ok := p.GetActiveList().PrevPageRequest(); ok {
				cmd = append(cmd, func() tea.Msg { return req })
			} else {
				cmd = append(cmd, p.GetActiveList().SetStatus("No previous page"))
			}
		case key.Matches(msg, mediaCenterKeyMap.Back):
			if p.lists.Len() > 1 {
				p.lists.Pop()
				cmd = append(cmd, p.GetActiveList().SetStatus("Back"))
			}
		}
	}
	return tea.Batch(cmd...)
}

func (p *panel) View() string {
	p.GetActiveList().SetSize(p.width, p.height)
	return p.GetActiveList().View()
}

func (p *panel) SetSize(width int, height int) {
	p.width = width
	p.height = height
}

func (p *panel) Prepare() tea.Cmd {
	if p.GetActiveList().state == initilized {
		requestCmd := tea.Cmd(func() tea.Msg {
			return MediaRequestForListKind(p.GetActiveList().kind)
		})
		return requestCmd
	}
	return nil
}

func (p *panel) SetStatus(s string) tea.Cmd {
	return p.GetActiveList().SetStatus(s)
}

func (p *panel) SetContent(entities []Entity, kind ListKind, pagination PaginationInfo, request MediaRequest) tea.Cmd {
	cmd := p.GetActiveList().SetContent(entities, kind)
	p.GetActiveList().ApplyPagination(pagination, request)
	p.GetActiveList().SetTitle(stackTitle(p.lists.Items))
	return cmd
}

func (p *panel) PrepareForKind(kind ListKind) {
	if kind == Tracks {
		return
	}
	p.lists.Push(newMediaList(kind))
}

func stackTitle(kinds []mediaList) string {
	if len(kinds) == 0 {
		return "Media"
	}
	if len(kinds) == 1 {
		return listTitle(kinds[0].kind)
	}

	parts := make([]string, 0, len(kinds))
	for i := 0; i < len(kinds)-1; i++ {
		parts = append(parts, listTitleAbbr(kinds[i].kind))
	}
	parts = append(parts, listTitle(kinds[len(kinds)-1].kind))
	return strings.Join(parts, ">")
}

func listTitleAbbr(kind ListKind) string {
	switch kind {
	case Albums:
		return "AL"
	case Artists:
		return "AR"
	case Playlists:
		return "PL"
	case Tracks:
		return "TR"
	case Shows:
		return "SH"
	case Episodes:
		return "EP"
	case AudioBooks:
		return "AB"
	default:
		return "Media"
	}
}
