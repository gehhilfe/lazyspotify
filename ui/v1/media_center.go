package v1

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/zmb3/spotify/v2"
)

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
)

type MediaRequest struct {
	kind   MediaRequestKind
	offset int
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
	Loading
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

func nextLibraryListKind(current ListKind) ListKind {
	switch current {
	case Playlists:
		return Tracks
	case Tracks:
		return Albums
	case Albums:
		return Artists
	default:
		return Playlists
	}
}

func MediaRequestForListKind(kind ListKind, offset int) MediaRequest {
	return MediaRequest{kind: requestKindForListKind(kind), offset: offset}
}

type MediaCenter struct {
	visibleList    mediaList
	cassettePlayer CassettePlayer
	displayScreen  displayScreen
}

func NewMediaCenter() MediaCenter {
	return MediaCenter{
		cassettePlayer: NewCassettePlayer(),
		visibleList:    newMediaList(),
		displayScreen:  newDisplayScreen(),
	}
}

func (m *MediaCenter) Update(msg tea.Msg) tea.Cmd {
	listCmd := m.visibleList.Update(msg)
	return listCmd
}
func (m *MediaCenter) StartLoading() tea.Cmd {
	return m.visibleList.StartLoading()
}

func (m *MediaCenter) SetContent(entities []Entity, kind ListKind) tea.Cmd {
	cmd := m.visibleList.SetContent(entities, kind)
	logger.Log.Info().Any("entities", entities).Int("kind", int(kind)).Msg("set content")
	logger.Log.Info().Int("kind", int(m.visibleList.kind)).Msg("set content visibleList")
	return cmd
}

func (m *MediaCenter) NextListKind() ListKind {
	return nextLibraryListKind(m.visibleList.kind)
}

func (e *Entity) Action(m *MediaCenter) tea.Cmd {
	return nil
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
