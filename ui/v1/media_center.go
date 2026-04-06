package v1

import (
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
	GetUserLibrary MediaRequestKind = iota
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

type MediaCenter struct {
	visibleList    mediaList
	cassettePlayer CassettePlayer
}

func NewMediaCenter() MediaCenter {
	return MediaCenter{
		cassettePlayer: NewCassettePlayer(),
		visibleList:    newMediaList(),
	}
}

func (m *MediaCenter) Update(msg tea.Msg) tea.Cmd {
	cmd := m.visibleList.Update(msg)
	return cmd
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

func (e *Entity) Action(m *MediaCenter) tea.Cmd {
	return nil
}

func AdaptSpotifyPlaylistPage(p *spotify.SimplePlaylistPage) []Entity {
	entities := make([]Entity, 0)
	logger.Log.Info().Any("p", p).Msg("AdaptSpotifyPlaylistPage")
	for _, pl := range p.Playlists {
		var img string
		if len(pl.Images) > 0 {
			img = pl.Images[0].URL
		}
		entities = append(entities,
			NewEntity(pl.Name,
				pl.Description,
				string(pl.ID),
				img,
			))
	}
	return entities
}
