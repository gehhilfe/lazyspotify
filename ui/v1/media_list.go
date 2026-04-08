package v1

import (
	"fmt"

	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/paginator"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
)

type state int

const (
	initilized state = iota
	loading
	ready
)

type mediaList struct {
	kind       ListKind
	items      []Entity
	list       list.Model
	pager      paginator.Model
	width      int
	height     int
	state      state
	pagination PaginationState
	request    MediaRequest
}

type PaginationState struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	HasNext     bool
	HasPrev     bool
	NextCursor  string
	PrevCursor  string
	history     []string
}

func newMediaList(kind ListKind) mediaList {
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
	listModel.Title = listTitle(kind)

	pager := paginator.New()
	pager.Type = paginator.Arabic
	pager.TotalPages = 1
	pager.Page = 0

	return mediaList{
		kind:       kind,
		list:       listModel,
		pager:      pager,
		state:      initilized,
		request:    MediaRequest{kind: requestKindForListKind(kind), page: 1, showLoading: true},
		pagination: PaginationState{CurrentPage: 1, history: []string{""}},
	}
}

func (m mediaList) View() string {
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
	if m.state == loading {
		m.list.Title = "Loading..."
		m.list.SetItems(nil)
	}
	if footer == "" {
		return m.list.View()
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.list.View(), footer)
}

func (m *mediaList) SetTitle(title string) {
	m.list.Title = title
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
	m.state = loading
	return tea.Batch(
		m.list.StartSpinner(),
	)
}

func (m *mediaList) StopLoading() {
	m.state = ready
	m.list.StopSpinner()
}

func (m *mediaList) SetContent(entities []Entity, kind ListKind) tea.Cmd {
	items := make([]list.Item, 0, len(entities))
	for _, entity := range entities {
		if entity.Name == "" {
			continue
		}
		items = append(items, mediaListItem{entity: entity})
	}
	m.kind = kind
	m.items = entities
	setItemsCmd := m.list.SetItems(items)
	logger.Log.Info().Any("items", entities).Int("kind", int(kind)).Msg("set content")
	m.StopLoading()
	return setItemsCmd
}

func (m *mediaList) ApplyPagination(info PaginationInfo, request MediaRequest) {
	if request.page <= 0 {
		request.page = 1
	}
	m.request.kind = request.kind
	m.kind = kindForRequestKind(request.kind)
	m.request.entityURI = request.entityURI
	m.request.showLoading = true

	m.pagination.CurrentPage = request.page
	m.pagination.TotalItems = info.TotalItems
	m.pagination.TotalPages = info.TotalPages
	m.pagination.HasNext = info.HasNext
	m.pagination.HasPrev = request.page > 1
	m.pagination.NextCursor = info.NextCursor

	if len(m.pagination.history) < request.page {
		missing := request.page - len(m.pagination.history)
		for i := 0; i < missing; i++ {
			m.pagination.history = append(m.pagination.history, "")
		}
	} else {
		m.pagination.history = m.pagination.history[:request.page]
	}
	m.pagination.history[request.page-1] = request.cursor

	m.pagination.PrevCursor = ""
	if request.page > 1 {
		m.pagination.PrevCursor = m.pagination.history[request.page-2]
	}

	totalPages := m.pagination.TotalPages
	if totalPages <= 0 {
		totalPages = 1
	}
	m.pager.TotalPages = totalPages
	m.pager.Page = max(0, m.pagination.CurrentPage-1)
	m.list.Title = listTitle(m.kind)
}

func (m *mediaList) NextPageRequest() (MediaRequest, bool) {
	if !m.pagination.HasNext || m.pagination.NextCursor == "" {
		return MediaRequest{}, false
	}
	return MediaRequest{
		kind:        m.request.kind,
		cursor:      m.pagination.NextCursor,
		page:        m.pagination.CurrentPage + 1,
		entityURI:   m.request.entityURI,
		showLoading: true,
	}, true
}

func (m *mediaList) PrevPageRequest() (MediaRequest, bool) {
	if !m.pagination.HasPrev {
		return MediaRequest{}, false
	}
	return MediaRequest{
		kind:        m.request.kind,
		cursor:      m.pagination.PrevCursor,
		page:        m.pagination.CurrentPage - 1,
		entityURI:   m.request.entityURI,
		showLoading: true,
	}, true
}

func (m *mediaList) StopSpinner() {
	m.list.StopSpinner()
}

func (m *mediaList) SetStatus(message string) tea.Cmd {
	return m.list.NewStatusMessage(message)
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
	default:
		return "Media"
	}
}

func kindForRequestKind(kind MediaRequestKind) ListKind {
	switch kind {
	case GetSavedTracks, GetPlaylistTracks, GetAlbumTracks:
		return Tracks
	case GetSavedAlbums, GetArtistAlbums:
		return Albums
	case GetFollowedArtists:
		return Artists
	default:
		return Playlists
	}
}
