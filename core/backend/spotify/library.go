package spotify

import (
	"context"
	"strings"

	"github.com/dubeyKartikay/lazyspotify/core/backend"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	coreplayer "github.com/dubeyKartikay/lazyspotify/core/player"
	"github.com/dubeyKartikay/lazyspotify/librespot/models"
	"github.com/dubeyKartikay/lazyspotify/spotify"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
	zspotify "github.com/zmb3/spotify/v2"
)

const (
	defaultPageSize    = 10
	albumTracksPageSize = 50
)

type Library struct {
	client *spotify.SpotifyClient
	player *coreplayer.Player
}

func NewLibrary(client *spotify.SpotifyClient, player *coreplayer.Player) *Library {
	return &Library{client: client, player: player}
}

func (l *Library) IsAuthError(err error) bool {
	return spotify.IsAuthError(err)
}

func (l *Library) GetUserPlaylists(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.GetUserPlaylists(ctx, offset)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptPlaylists(page)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func (l *Library) GetSavedTracks(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.GetSavedTracks(ctx, offset)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptSavedTracks(page)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func (l *Library) GetSavedAlbums(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.GetSavedAlbums(ctx, offset)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptSavedAlbums(page)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func (l *Library) GetFollowedArtists(ctx context.Context, cursor string, page int) ([]common.Entity, common.PaginationInfo, error) {
	result, err := l.client.GetFollowedArtists(ctx, cursor)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptArtistsPage(result.Artists)
	return entities, backend.PaginationFromCursor(page, len(entities), int(result.Total), defaultPageSize, result.Cursor.After), nil
}

func (l *Library) GetPlaylistTracks(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	if l.player == nil {
		return nil, common.PaginationInfo{}, nil
	}
	offset := backend.DecodeOffsetCursor(cursor)
	resp, err := l.player.GetPlaylistTracks(ctx, uri, offset, defaultPageSize)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptResolvedTracks(resp.Tracks)
	nextCursor := ""
	if resp.HasNext {
		nextCursor = backend.EncodeOffsetCursor(offset + defaultPageSize)
	}
	currentPage := 1
	if defaultPageSize > 0 {
		currentPage = (offset / defaultPageSize) + 1
	}
	return entities, common.PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  backend.TotalPages(resp.Total, defaultPageSize),
		TotalItems:  resp.Total,
		HasNext:     resp.HasNext,
		NextCursor:  nextCursor,
	}, nil
}

func (l *Library) GetArtistAlbums(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.GetArtistAlbums(ctx, uri, offset)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptSimpleAlbums(page.Albums)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func (l *Library) GetAlbumTracks(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.GetAlbumTracks(ctx, uri, offset)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptSimpleTracks(page.Tracks)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), albumTracksPageSize), nil
}

func (l *Library) SearchPlaylists(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.SearchPlaylists(ctx, query, offset, defaultPageSize)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptPlaylists(page)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func (l *Library) SearchTracks(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.SearchTracks(ctx, query, offset, defaultPageSize)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptFullTracks(page.Tracks)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func (l *Library) SearchAlbums(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.SearchAlbums(ctx, query, offset, defaultPageSize)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptSimpleAlbums(page.Albums)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func (l *Library) SearchArtists(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	page, err := l.client.SearchArtists(ctx, query, offset, defaultPageSize)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := adaptArtistsPage(page.Artists)
	return entities, backend.PaginationFromOffset(offset, len(entities), int(page.Total), defaultPageSize), nil
}

func adaptPlaylists(page *zspotify.SimplePlaylistPage) []common.Entity {
	logger.Log.Info().Any("p", page).Msg("adapt spotify playlist page")
	return common.MapSlice(page.Playlists, func(pl zspotify.SimplePlaylist) common.Entity {
		return common.NewEntity(pl.Name, pl.Description, string(pl.URI), imageURL(pl.Images))
	})
}

func adaptSavedTracks(page *zspotify.SavedTrackPage) []common.Entity {
	return common.MapSlice(page.Tracks, func(savedTrack zspotify.SavedTrack) common.Entity {
		track := savedTrack.FullTrack
		desc := strings.TrimSpace(joinArtists(track.Artists))
		if track.Album.Name != "" {
			if desc != "" {
				desc += " • " + track.Album.Name
			} else {
				desc = track.Album.Name
			}
		}
		return common.NewEntity(track.Name, desc, string(track.URI), imageURL(track.Album.Images))
	})
}

func adaptSavedAlbums(page *zspotify.SavedAlbumPage) []common.Entity {
	return common.MapSlice(page.Albums, func(savedAlbum zspotify.SavedAlbum) common.Entity {
		album := savedAlbum.FullAlbum
		return common.NewEntity(album.Name, joinArtists(album.Artists), string(album.URI), imageURL(album.Images))
	})
}

func adaptArtistsPage(artists []zspotify.FullArtist) []common.Entity {
	return common.MapSlice(artists, func(artist zspotify.FullArtist) common.Entity {
		desc := ""
		if len(artist.Genres) > 0 {
			desc = strings.Join(artist.Genres, ", ")
		}
		return common.NewEntity(artist.Name, desc, string(artist.URI), imageURL(artist.Images))
	})
}

func adaptFullTracks(tracks []zspotify.FullTrack) []common.Entity {
	return common.MapSlice(tracks, func(track zspotify.FullTrack) common.Entity {
		desc := strings.TrimSpace(joinArtists(track.Artists))
		if track.Album.Name != "" {
			if desc != "" {
				desc += " • " + track.Album.Name
			} else {
				desc = track.Album.Name
			}
		}
		return common.NewEntity(track.Name, desc, string(track.URI), imageURL(track.Album.Images))
	})
}

func adaptResolvedTracks(tracks []models.ResolvedTrack) []common.Entity {
	return common.MapSlice(tracks, func(track models.ResolvedTrack) common.Entity {
		desc := strings.TrimSpace(strings.Join(track.Artists, ", "))
		if track.AlbumName != "" {
			if desc != "" {
				desc += " • " + track.AlbumName
			} else {
				desc = track.AlbumName
			}
		}
		return common.NewEntity(track.Name, desc, track.URI, track.Img)
	})
}

func adaptSimpleAlbums(albums []zspotify.SimpleAlbum) []common.Entity {
	return common.MapSlice(albums, func(album zspotify.SimpleAlbum) common.Entity {
		return common.NewEntity(album.Name, joinArtists(album.Artists), string(album.URI), imageURL(album.Images))
	})
}

func adaptSimpleTracks(tracks []zspotify.SimpleTrack) []common.Entity {
	return common.MapSlice(tracks, func(track zspotify.SimpleTrack) common.Entity {
		return common.NewEntity(track.Name, strings.TrimSpace(joinArtists(track.Artists)), string(track.URI), "")
	})
}

func imageURL(images []zspotify.Image) string {
	if len(images) == 0 {
		return ""
	}
	return images[0].URL
}

func joinArtists(artists []zspotify.SimpleArtist) string {
	if len(artists) == 0 {
		return ""
	}
	names := common.MapSlice(artists, func(artist zspotify.SimpleArtist) string {
		return artist.Name
	})
	return strings.Join(names, ", ")
}
