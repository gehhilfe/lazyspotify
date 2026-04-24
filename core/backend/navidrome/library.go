package navidrome

import (
	"context"
	"fmt"
	"strings"

	"github.com/dubeyKartikay/lazyspotify/core/backend"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
)

const (
	defaultPageSize = 10
	uriPrefix       = "nd:"
)

type Library struct {
	client *Client
}

func NewLibrary(client *Client) *Library {
	return &Library{client: client}
}

func (l *Library) IsAuthError(err error) bool {
	return IsAuthError(err)
}

func TrackURI(id string) string    { return uriPrefix + "track:" + id }
func AlbumURI(id string) string    { return uriPrefix + "album:" + id }
func ArtistURI(id string) string   { return uriPrefix + "artist:" + id }
func PlaylistURI(id string) string { return uriPrefix + "playlist:" + id }

// ParseURI returns (kind, id) for any nd: URI, or empty strings if it doesn't match.
func ParseURI(uri string) (kind, id string) {
	if !strings.HasPrefix(uri, uriPrefix) {
		return "", ""
	}
	rest := strings.TrimPrefix(uri, uriPrefix)
	k, i, ok := strings.Cut(rest, ":")
	if !ok {
		return "", ""
	}
	return k, i
}

func idForKind(uri, wantKind string) (string, error) {
	kind, id := ParseURI(uri)
	if kind != wantKind || id == "" {
		return "", fmt.Errorf("invalid navidrome uri %q: expected %s", uri, wantKind)
	}
	return id, nil
}

func paginateClientSide[T any](items []T, offset, size int) (page []T, total int, nextOffset int, hasNext bool) {
	total = len(items)
	if offset < 0 {
		offset = 0
	}
	if offset >= total {
		return nil, total, 0, false
	}
	end := min(offset+size, total)
	return items[offset:end], total, end, end < total
}

func (l *Library) GetUserPlaylists(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	playlists, err := l.client.GetPlaylists(ctx)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	page, total, _, _ := paginateClientSide(playlists, offset, defaultPageSize)
	entities := common.MapSlice(page, l.adaptPlaylist)
	return entities, backend.PaginationFromOffset(offset, len(entities), total, defaultPageSize), nil
}

func (l *Library) GetSavedTracks(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	songs, _, _, err := l.client.GetStarred2(ctx)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	page, total, _, _ := paginateClientSide(songs, offset, defaultPageSize)
	entities := common.MapSlice(page, l.adaptSong)
	return entities, backend.PaginationFromOffset(offset, len(entities), total, defaultPageSize), nil
}

func (l *Library) GetSavedAlbums(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	albums, err := l.client.GetAlbumList2(ctx, "alphabeticalByName", offset, defaultPageSize)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := common.MapSlice(albums, l.adaptAlbum)
	hasNext := len(entities) == defaultPageSize
	total := offset + len(entities)
	if hasNext {
		total = offset + len(entities) + 1
	}
	nextCursor := ""
	if hasNext {
		nextCursor = backend.EncodeOffsetCursor(offset + defaultPageSize)
	}
	currentPage := (offset / defaultPageSize) + 1
	return entities, common.PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  backend.TotalPages(total, defaultPageSize),
		TotalItems:  total,
		HasNext:     hasNext,
		NextCursor:  nextCursor,
	}, nil
}

func (l *Library) GetFollowedArtists(ctx context.Context, cursor string, _ int) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	artists, err := l.client.GetArtists(ctx)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	page, total, _, _ := paginateClientSide(artists, offset, defaultPageSize)
	entities := common.MapSlice(page, l.adaptArtist)
	return entities, backend.PaginationFromOffset(offset, len(entities), total, defaultPageSize), nil
}

func (l *Library) GetPlaylistTracks(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	id, err := idForKind(uri, "playlist")
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	offset := backend.DecodeOffsetCursor(cursor)
	pl, err := l.client.GetPlaylist(ctx, id)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	page, total, _, _ := paginateClientSide(pl.Entry, offset, defaultPageSize)
	entities := common.MapSlice(page, l.adaptSong)
	return entities, backend.PaginationFromOffset(offset, len(entities), total, defaultPageSize), nil
}

func (l *Library) GetArtistAlbums(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	id, err := idForKind(uri, "artist")
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	offset := backend.DecodeOffsetCursor(cursor)
	artist, err := l.client.GetArtist(ctx, id)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	page, total, _, _ := paginateClientSide(artist.Album, offset, defaultPageSize)
	entities := common.MapSlice(page, l.adaptAlbum)
	return entities, backend.PaginationFromOffset(offset, len(entities), total, defaultPageSize), nil
}

func (l *Library) GetAlbumTracks(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	id, err := idForKind(uri, "album")
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	offset := backend.DecodeOffsetCursor(cursor)
	al, err := l.client.GetAlbum(ctx, id)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	page, total, _, _ := paginateClientSide(al.Song, offset, defaultPageSize)
	entities := common.MapSlice(page, l.adaptSong)
	return entities, backend.PaginationFromOffset(offset, len(entities), total, defaultPageSize), nil
}

func (l *Library) SearchPlaylists(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	playlists, err := l.client.GetPlaylists(ctx)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	q := strings.ToLower(strings.TrimSpace(query))
	matched := playlists[:0]
	for _, pl := range playlists {
		if q == "" || strings.Contains(strings.ToLower(pl.Name), q) {
			matched = append(matched, pl)
		}
	}
	page, total, _, _ := paginateClientSide(matched, offset, defaultPageSize)
	entities := common.MapSlice(page, l.adaptPlaylist)
	return entities, backend.PaginationFromOffset(offset, len(entities), total, defaultPageSize), nil
}

func (l *Library) SearchTracks(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	res, err := l.client.Search3(ctx, query, defaultPageSize, offset, 0, 0, 0, 0)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := common.MapSlice(res.Songs, l.adaptSong)
	return entities, searchPagination(offset, len(entities)), nil
}

func (l *Library) SearchAlbums(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	res, err := l.client.Search3(ctx, query, 0, 0, defaultPageSize, offset, 0, 0)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := common.MapSlice(res.Albums, l.adaptAlbum)
	return entities, searchPagination(offset, len(entities)), nil
}

func (l *Library) SearchArtists(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error) {
	offset := backend.DecodeOffsetCursor(cursor)
	res, err := l.client.Search3(ctx, query, 0, 0, 0, 0, defaultPageSize, offset)
	if err != nil {
		return nil, common.PaginationInfo{}, err
	}
	entities := common.MapSlice(res.Artists, l.adaptArtist)
	return entities, searchPagination(offset, len(entities)), nil
}

// searchPagination infers a usable paging state without a reliable total.
// Subsonic's search3 returns "up to count" items — we treat a full page as "more may follow".
func searchPagination(offset, got int) common.PaginationInfo {
	hasNext := got == defaultPageSize
	total := offset + got
	if hasNext {
		total = offset + got + 1
	}
	nextCursor := ""
	if hasNext {
		nextCursor = backend.EncodeOffsetCursor(offset + defaultPageSize)
	}
	return common.PaginationInfo{
		CurrentPage: (offset / defaultPageSize) + 1,
		TotalPages:  backend.TotalPages(total, defaultPageSize),
		TotalItems:  total,
		HasNext:     hasNext,
		NextCursor:  nextCursor,
	}
}

// --- adapters ---

func (l *Library) adaptPlaylist(p Playlist) common.Entity {
	desc := fmt.Sprintf("%d tracks", p.SongCount)
	if p.Comment != "" {
		desc = p.Comment
	}
	return common.NewEntity(p.Name, desc, PlaylistURI(p.ID), l.client.CoverArtURL(p.CoverArt))
}

func (l *Library) adaptSong(s Song) common.Entity {
	desc := s.Artist
	if s.Album != "" {
		if desc != "" {
			desc += " • " + s.Album
		} else {
			desc = s.Album
		}
	}
	return common.NewEntity(s.Title, desc, TrackURI(s.ID), l.client.CoverArtURL(s.CoverArt))
}

func (l *Library) adaptAlbum(a Album) common.Entity {
	desc := a.Artist
	return common.NewEntity(a.Name, desc, AlbumURI(a.ID), l.client.CoverArtURL(a.CoverArt))
}

func (l *Library) adaptArtist(a Artist) common.Entity {
	desc := ""
	if a.AlbumCount > 0 {
		desc = fmt.Sprintf("%d albums", a.AlbumCount)
	}
	return common.NewEntity(a.Name, desc, ArtistURI(a.ID), l.client.CoverArtURL(a.CoverArt))
}
