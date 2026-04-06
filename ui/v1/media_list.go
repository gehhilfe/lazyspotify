package v1

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
)

type mediaList struct {
	kind   ListKind
	items  []Entity
	list   list.Model
	styles styles
	width  int
	height int
}

type styles struct {
	panel          lipgloss.Style
	panelNav       lipgloss.Style
	panelNavActive lipgloss.Style
	panelNavMuted  lipgloss.Style
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

func newMediaList() mediaList {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.
		Foreground(lipgloss.Color("252")).
		PaddingLeft(1)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.
		Foreground(lipgloss.Color("245")).
		PaddingLeft(1)
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("14")).
		Bold(true).
		BorderLeft(false).
		PaddingLeft(1)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("117")).
		BorderLeft(false).
		PaddingLeft(1)
	delegate.SetSpacing(1)
	listModel := list.New(nil, delegate, 0, 0)
	
	styles := listModel.Styles
	styles.Title = styles.Title.MarginLeft(1)

	styles.TitleBar = lipgloss.NewStyle().MarginBottom(1)
	styles.NoItems = styles.NoItems.Foreground(lipgloss.Color("8"))
	listModel.Styles = styles
	listModel.SetShowHelp(false)
	listModel.SetShowStatusBar(false)
	listModel.SetShowFilter(false)
	listModel.SetShowPagination(false)
	listModel.InfiniteScrolling = true
	listModel.Title = listTitle(Loading)

	return mediaList{
		kind:   Loading,
		list:   listModel,
		styles: defaultStyles(),
	}
}

func (m mediaList) View() string {
	panel := m.styles.panel.Width(m.width).Height(m.height).Render("")
	panelNav := m.renderPanelNav()
	listWidth := m.width - 4
	listHeight := m.height - 2
	m.list.SetSize(listWidth, listHeight)
	panelNavX := (m.width - lipgloss.Width(panelNav)) / 2
	if panelNavX < 1 {
		panelNavX = 1
	}

	layers := []*lipgloss.Layer{
		lipgloss.NewLayer(panel).ID("panel"),
		lipgloss.NewLayer(panelNav).X(panelNavX).Y(0).ID("panel-nav"),
		lipgloss.NewLayer(m.list.View()).X(1).Y(1).ID("list"),
	}
	compositor := lipgloss.NewCompositor(layers...)
	return compositor.Render()
}

func (m mediaList) renderPanelNav() string {
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
		if m.kind == segment.kind {
			parts = append(parts, m.styles.panelNavActive.Render(segment.label))
			continue
		}
		parts = append(parts, m.styles.panelNavMuted.Render(segment.label))
	}

	return m.styles.panelNav.Render(strings.Join(parts, " - "))
}
func (m *mediaList) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return cmd
}

func (m *mediaList) SetSize(width, height int) {
	m.width = width
	m.height = height

	listStyles := m.list.Styles
	m.list.Styles = listStyles
}
func (m *mediaList) StartLoading() tea.Cmd {
	m.kind = Loading
	m.items = nil
	m.list.Title = listTitle(Loading)
	m.list.SetItems(nil)
	return tea.Batch(
		m.list.StartSpinner(),
	)
}

func (m *mediaList) SetContent(entities []Entity, kind ListKind) tea.Cmd {
	items := make([]list.Item, 0, len(entities))
	for _, entity := range entities {
		if(entity.Name == "") {
			continue
		}
		items = append(items, mediaListItem{entity: entity})
	}
	m.kind = kind
	m.items = entities
	setItemsCmd := m.list.SetItems(items)
	logger.Log.Info().Any("items", entities).Int("kind", int(kind)).Msg("set content")
	m.list.StopSpinner()
	m.list.Title = listTitle(kind)
	return setItemsCmd
}

type mediaListItem struct {
	entity Entity
}

func (i mediaListItem) Title() string {
	return i.entity.Name
}

func (i mediaListItem) Description() string {
	return i.entity.Desc
}

func (i mediaListItem) FilterValue() string {
	return fmt.Sprintf("%s %s", i.entity.Name, i.entity.Desc)
}

func listTitle(kind ListKind) string {
	switch kind {
	case Albums:
		return "Albums"
	case Artists:
		return "Artists"
	case Playlists:
		return "Playlists"
	case Tracks:
		return "Tracks"
	case Shows:
		return "Shows"
	case Episodes:
		return "Episodes"
	case AudioBooks:
		return "Audiobooks"
	case Loading:
		return "Loading"
	default:
		return "Media"
	}
}
