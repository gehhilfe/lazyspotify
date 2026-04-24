package app

import (
	"context"

	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
)

func (m *Model) handleGetUserPlaylists(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.GetUserPlaylists(context.Background(), request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Playlists, pagination: pagination, request: request}
	}
}

func (m *Model) handleGetSavedTracks(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.GetSavedTracks(context.Background(), request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Tracks, pagination: pagination, request: request}
	}
}

func (m *Model) handleGetSavedAlbums(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.GetSavedAlbums(context.Background(), request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Albums, pagination: pagination, request: request}
	}
}

func (m *Model) handleGetFollowedArtists(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.GetFollowedArtists(context.Background(), request.Cursor, request.Page)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Artists, pagination: pagination, request: request}
	}
}

func (m *Model) handleSearchPlaylists(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.SearchPlaylists(context.Background(), request.Query, request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Playlists, pagination: pagination, request: request}
	}
}

func (m *Model) handleSearchTracks(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.SearchTracks(context.Background(), request.Query, request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Tracks, pagination: pagination, request: request}
	}
}

func (m *Model) handleSearchAlbums(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.SearchAlbums(context.Background(), request.Query, request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Albums, pagination: pagination, request: request}
	}
}

func (m *Model) handleSearchArtists(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.SearchArtists(context.Background(), request.Query, request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Artists, pagination: pagination, request: request}
	}
}

func (m *Model) handleGetPlaylistTracks(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.GetPlaylistTracks(context.Background(), request.EntityURI, request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Tracks, pagination: pagination, request: request}
	}
}

func (m *Model) handleGetArtistAlbums(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.GetArtistAlbums(context.Background(), request.EntityURI, request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Albums, pagination: pagination, request: request}
	}
}

func (m *Model) handleGetAlbumTracks(request common.MediaRequest) tea.Cmd {
	if m.library == nil {
		return nil
	}
	return func() tea.Msg {
		entities, pagination, err := m.library.GetAlbumTracks(context.Background(), request.EntityURI, request.Cursor)
		if err != nil {
			return mediaLoadErrMsg{err: err, request: request}
		}
		return mediaLoadedMsg{entities: entities, kind: common.Tracks, pagination: pagination, request: request}
	}
}

func (m *Model) handlePlayTrackRequest(request common.MediaRequest) tea.Cmd {
	if m.player == nil {
		return m.mediaCenter.SetStatus(request.PanelKind, "Player not ready")
	}
	m.mediaCenter.SetDisplay("Loading...")
	m.playerReady = false
	m.playing = false
	m.mediaCenter.CloseLibrary()
	return func() tea.Msg {
		err := m.player.PlayTrack(context.Background(), request.EntityURI, request.ContextURI)
		if err != nil {
			return playTrackErrMsg{err: err, panelKind: request.PanelKind}
		}
		return playTrackOkMsg{panelKind: request.PanelKind}
	}
}
