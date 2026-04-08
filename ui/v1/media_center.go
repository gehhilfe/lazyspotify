package v1

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/dubeyKartikay/lazyspotify/librespot/models"
	"github.com/zmb3/spotify/v2"
)

var mediaCenterKeyMap = struct {
	Select      key.Binding
	TogglePanel key.Binding
	Back        key.Binding
	NextPanel   key.Binding
	NextPage    key.Binding
	PrevPage    key.Binding
}{
	Select:      key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	TogglePanel: key.NewBinding(key.WithKeys("P"), key.WithHelp("P", "toggle panel")),
	Back:        key.NewBinding(key.WithKeys("backspace", "delete"), key.WithHelp("del", "back")),
	NextPanel:   key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next panel")),
	NextPage:    key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl+d", "next page")),
	PrevPage:    key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "prev page")),
}

type Entity struct {
	Name string
	Desc string
	ID   string
	Img  string
}

type MediaRequestKind int

const (
	GetUserPlaylists MediaRequestKind = iota
	GetSavedTracks
	GetSavedAlbums
	GetFollowedArtists
	GetPlaylistTracks
	GetArtistAlbums
	GetAlbumTracks
	PlayTrack
)

type MediaRequest struct {
	kind        MediaRequestKind
	cursor      string
	page        int
	entityURI   string
	contextURI  string
	showLoading bool
}

type PaginationInfo struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	HasNext     bool
	NextCursor  string
}

func NewEntity(name string, desc string, uri string, img string) Entity {
	return Entity{
		Name: name,
		Desc: desc,
		ID:   uri,
		Img:  img,
	}
}

type ListKind int

const (
	Playlists ListKind = iota
	Albums
	Artists
	Tracks
	Shows
	Episodes
	AudioBooks
)

func requestKindForListKind(kind ListKind) MediaRequestKind {
	switch kind {
	case Tracks:
		return GetSavedTracks
	case Albums:
		return GetSavedAlbums
	case Artists:
		return GetFollowedArtists
	default:
		return GetUserPlaylists
	}
}

func MediaRequestForListKind(kind ListKind) MediaRequest {
	return MediaRequest{kind: requestKindForListKind(kind), page: 1, showLoading: true}
}

type MediaCenter struct {
	mediaPanel     MediaPanel
	cassettePlayer CassettePlayer
	displayScreen  displayScreen
	mediaListOpen  bool
}

func NewMediaCenter() MediaCenter {
	m := MediaCenter{
		mediaPanel:     NewMediaPanel(),
		cassettePlayer: NewCassettePlayer(),
		displayScreen:  newDisplayScreen(),
	}
	return m
}

func (m *MediaCenter) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, mediaCenterKeyMap.TogglePanel):
			m.mediaListOpen = !m.mediaListOpen
			return nil
		}
	}
	cmd := m.mediaPanel.Update(msg)
	return cmd
}
func (m *MediaCenter) StartLoading() tea.Cmd {
	return m.mediaPanel.StartLoading()
}

func (m *MediaCenter) SetContent(entities []Entity, kind ListKind, pagination PaginationInfo, request MediaRequest) tea.Cmd {
	cmd := m.mediaPanel.SetContent(entities, kind, pagination, request)
	logger.Log.Info().Any("entities", entities).Int("kind", int(kind)).Msg("set content")
	return cmd
}

func (m *MediaCenter) SetStatus(message string) tea.Cmd {
	return m.mediaPanel.SetStatus(message)
}

func (e *Entity) Action(p *panel) tea.Cmd {
	kind := p.GetActiveList().kind
	p.PrepareForKind(kind)
	switch kind {
	case Playlists:
		return func() tea.Msg {
			return MediaRequest{
				kind:        GetPlaylistTracks,
				page:        1,
				entityURI:   e.ID,
				showLoading: true,
			}
		}
	case Artists:
		return func() tea.Msg {
			return MediaRequest{
				kind:        GetArtistAlbums,
				page:        1,
				entityURI:   e.ID,
				showLoading: true,
			}
		}
	case Albums:
		return func() tea.Msg {
			return MediaRequest{
				kind:        GetAlbumTracks,
				page:        1,
				entityURI:   e.ID,
				showLoading: true,
			}
		}
	case Tracks:
		return func() tea.Msg {
			return MediaRequest{
				kind:        PlayTrack,
				entityURI:   e.ID,
				contextURI:  p.GetActiveList().request.entityURI,
				showLoading: false,
			}
		}
	default:
		return nil
	}
}

func AdaptSpotifyPlaylistPage(p *spotify.SimplePlaylistPage) []Entity {
	entities := make([]Entity, 0)
	logger.Log.Info().Any("p", p).Msg("AdaptSpotifyPlaylistPage")
	for _, pl := range p.Playlists {
		img := imageURL(pl.Images)
		entities = append(entities,
			NewEntity(pl.Name,
				pl.Description,
				string(pl.URI),
				img,
			))
	}
	return entities
}

func AdaptSpotifySavedTrackPage(p *spotify.SavedTrackPage) []Entity {
	entities := make([]Entity, 0, len(p.Tracks))
	for _, savedTrack := range p.Tracks {
		track := savedTrack.FullTrack
		desc := strings.TrimSpace(joinArtists(track.Artists))
		if track.Album.Name != "" {
			if desc != "" {
				desc += " • " + track.Album.Name
			} else {
				desc = track.Album.Name
			}
		}
		entities = append(entities,
			NewEntity(track.Name,
				desc,
				string(track.URI),
				imageURL(track.Album.Images),
			))
	}
	return entities
}

func AdaptSpotifySavedAlbumPage(p *spotify.SavedAlbumPage) []Entity {
	entities := make([]Entity, 0, len(p.Albums))
	for _, savedAlbum := range p.Albums {
		album := savedAlbum.FullAlbum
		entities = append(entities,
			NewEntity(album.Name,
				joinArtists(album.Artists),
				string(album.URI),
				imageURL(album.Images),
			))
	}
	return entities
}

func AdaptSpotifyFollowedArtistsPage(p *spotify.FullArtistCursorPage) []Entity {
	entities := make([]Entity, 0, len(p.Artists))
	for _, artist := range p.Artists {
		desc := ""
		if len(artist.Genres) > 0 {
			desc = strings.Join(artist.Genres, ", ")
		}
		entities = append(entities,
			NewEntity(artist.Name,
				desc,
				string(artist.URI),
				imageURL(artist.Images),
			))
	}
	return entities
}

func AdaptSpotifyPlaylistTracks(tracks []spotify.FullTrack) []Entity {
	entities := make([]Entity, 0, len(tracks))
	for _, track := range tracks {
		desc := strings.TrimSpace(joinArtists(track.Artists))
		if track.Album.Name != "" {
			if desc != "" {
				desc += " • " + track.Album.Name
			} else {
				desc = track.Album.Name
			}
		}
		entities = append(entities, NewEntity(track.Name, desc, string(track.URI), imageURL(track.Album.Images)))
	}
	return entities
}

func AdaptResolvedPlaylistTracks(tracks []models.ResolvedTrack) []Entity {
	entities := make([]Entity, 0, len(tracks))
	for _, track := range tracks {
		desc := strings.TrimSpace(strings.Join(track.Artists, ", "))
		if track.AlbumName != "" {
			if desc != "" {
				desc += " • " + track.AlbumName
			} else {
				desc = track.AlbumName
			}
		}
		entities = append(entities, NewEntity(track.Name, desc, track.URI, track.Img))
	}
	return entities
}

func AdaptSpotifyArtistAlbums(albums []spotify.SimpleAlbum) []Entity {
	entities := make([]Entity, 0, len(albums))
	for _, album := range albums {
		entities = append(entities,
			NewEntity(album.Name,
				joinArtists(album.Artists),
				string(album.URI),
				imageURL(album.Images),
			))
	}
	return entities
}

func AdaptSpotifyAlbumTracks(tracks []spotify.SimpleTrack) []Entity {
	entities := make([]Entity, 0, len(tracks))
	for _, track := range tracks {
		desc := strings.TrimSpace(joinArtists(track.Artists))
		entities = append(entities,
			NewEntity(track.Name,
				desc,
				string(track.URI),
				"",
			))
	}
	return entities
}

func imageURL(images []spotify.Image) string {
	if len(images) == 0 {
		return ""
	}
	return images[0].URL
}

func joinArtists(artists []spotify.SimpleArtist) string {
	if len(artists) == 0 {
		return ""
	}
	names := make([]string, 0, len(artists))
	for _, artist := range artists {
		names = append(names, artist.Name)
	}
	return strings.Join(names, ", ")
}
