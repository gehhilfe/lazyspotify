package mediapanel

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/medialist"
)

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	panel := m.activePanel()
	cmd := panel.Update(msg)
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, common.MediaCenterKeyMap.NextPanel) {
			cmd = tea.Batch(cmd, m.activateNextPanel())
		}
	}
	return cmd
}

func (m *Model) StartLoading() tea.Cmd {
	return m.activePanel().activeList().StartLoading()
}

func (m *Model) SetStatus(message string) tea.Cmd {
	return m.activePanel().SetStatus(message)
}

func (m *Model) SetContent(entities []common.Entity, kind common.ListKind, pagination common.PaginationInfo, request common.MediaRequest) tea.Cmd {
	return m.activePanel().SetContent(entities, kind, pagination, request)
}

func (m *Model) activateNextPanel() tea.Cmd {
	m.active = (m.active + 1) % len(m.panels)
	return m.activePanel().Prepare()
}

func (p *panel) activeList() *medialist.Model {
	return p.lists.Peek()
}

func (p *panel) Update(msg tea.Msg) tea.Cmd {
	cmds := []tea.Cmd{p.activeList().Update(msg)}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, common.MediaCenterKeyMap.Select):
			if req, ok := p.selectedAction(); ok {
				cmds = append(cmds, func() tea.Msg { return req })
			}
		case key.Matches(msg, common.MediaCenterKeyMap.NextPage):
			if req, ok := p.activeList().NextPageRequest(); ok {
				cmds = append(cmds, func() tea.Msg { return req })
			} else {
				cmds = append(cmds, p.activeList().SetStatus("No next page"))
			}
		case key.Matches(msg, common.MediaCenterKeyMap.PrevPage):
			if req, ok := p.activeList().PrevPageRequest(); ok {
				cmds = append(cmds, func() tea.Msg { return req })
			} else {
				cmds = append(cmds, p.activeList().SetStatus("No previous page"))
			}
		case key.Matches(msg, common.MediaCenterKeyMap.Back):
			if p.lists.Len() > 1 {
				p.lists.Pop()
				cmds = append(cmds, p.activeList().SetStatus("Back"))
			}
		}
	}
	return tea.Batch(cmds...)
}

func (p *panel) Prepare() tea.Cmd {
	if p.activeList().State() == medialist.Initialized {
		request := common.MediaRequestForListKind(p.activeList().Kind())
		return func() tea.Msg { return request }
	}
	return nil
}

func (p *panel) SetStatus(message string) tea.Cmd {
	return p.activeList().SetStatus(message)
}

func (p *panel) SetContent(entities []common.Entity, kind common.ListKind, pagination common.PaginationInfo, request common.MediaRequest) tea.Cmd {
	cmd := p.activeList().SetContent(entities, kind)
	p.activeList().ApplyPagination(pagination, request)
	p.activeList().SetTitle(stackTitle(p.lists.Items))
	return cmd
}

func (p *panel) selectedAction() (common.MediaRequest, bool) {
	entity, ok := p.activeList().SelectedEntity()
	if !ok {
		return common.MediaRequest{}, false
	}
	kind := p.activeList().Kind()
	p.prepareForKind(kind)
	switch kind {
	case common.Playlists:
		return common.MediaRequest{Kind: common.GetPlaylistTracks, Page: 1, EntityURI: entity.ID, ShowLoading: true}, true
	case common.Artists:
		return common.MediaRequest{Kind: common.GetArtistAlbums, Page: 1, EntityURI: entity.ID, ShowLoading: true}, true
	case common.Albums:
		return common.MediaRequest{Kind: common.GetAlbumTracks, Page: 1, EntityURI: entity.ID, ShowLoading: true}, true
	case common.Tracks:
		return common.MediaRequest{
			Kind:        common.PlayTrack,
			EntityURI:   entity.ID,
			ContextURI:  p.activeList().Request().EntityURI,
			ShowLoading: false,
		}, true
	default:
		return common.MediaRequest{}, false
	}
}

func (p *panel) prepareForKind(kind common.ListKind) {
	if kind == common.Tracks {
		return
	}
	p.lists.Push(medialist.NewModel(kind))
}

func stackTitle(lists []medialist.Model) string {
	if len(lists) == 0 {
		return "Media"
	}
	if len(lists) == 1 {
		return common.ListTitle(lists[0].Kind())
	}

	parts := make([]string, 0, len(lists))
	for i := 0; i < len(lists)-1; i++ {
		parts = append(parts, common.ListTitleAbbr(lists[i].Kind()))
	}
	parts = append(parts, common.ListTitle(lists[len(lists)-1].Kind()))
	return strings.Join(parts, ">")
}
