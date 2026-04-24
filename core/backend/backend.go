package backend

import (
	"context"
	"strconv"

	"github.com/dubeyKartikay/lazyspotify/librespot/models"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
)

type Library interface {
	GetUserPlaylists(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error)
	GetSavedTracks(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error)
	GetSavedAlbums(ctx context.Context, cursor string) ([]common.Entity, common.PaginationInfo, error)
	GetFollowedArtists(ctx context.Context, cursor string, page int) ([]common.Entity, common.PaginationInfo, error)
	GetPlaylistTracks(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error)
	GetArtistAlbums(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error)
	GetAlbumTracks(ctx context.Context, uri, cursor string) ([]common.Entity, common.PaginationInfo, error)
	SearchPlaylists(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error)
	SearchTracks(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error)
	SearchAlbums(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error)
	SearchArtists(ctx context.Context, query, cursor string) ([]common.Entity, common.PaginationInfo, error)
	IsAuthError(err error) bool
}

type Player interface {
	PlayTrack(ctx context.Context, uri, contextURI string) error
	PlayPause(ctx context.Context) error
	Seek(ctx context.Context, position int, relative bool) error
	Next(ctx context.Context) error
	Previous(ctx context.Context) error
	SetVolume(ctx context.Context, volume int, relative bool) error
	GetVolume(ctx context.Context) (common.VolumeInfo, error)
	Start(ctx context.Context) error
	WaitTillReady() error
	WaitForDaemonFailure() error
	Destroy(ctx context.Context)
	Events() <-chan models.PlayerEvent
}

func DecodeOffsetCursor(cursor string) int {
	if cursor == "" {
		return 0
	}
	value, err := strconv.Atoi(cursor)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func EncodeOffsetCursor(offset int) string {
	if offset < 0 {
		offset = 0
	}
	return strconv.Itoa(offset)
}

func TotalPages(totalItems, pageSize int) int {
	if totalItems <= 0 || pageSize <= 0 {
		return 1
	}
	pages := totalItems / pageSize
	if totalItems%pageSize != 0 {
		pages++
	}
	if pages <= 0 {
		return 1
	}
	return pages
}

func PaginationFromOffset(offset, count, total, pageSize int) common.PaginationInfo {
	currentPage := 1
	if pageSize > 0 {
		currentPage = (offset / pageSize) + 1
	}
	hasNext := offset+count < total
	nextCursor := ""
	if hasNext {
		nextCursor = EncodeOffsetCursor(offset + pageSize)
	}
	return common.PaginationInfo{
		CurrentPage: currentPage,
		TotalPages:  TotalPages(total, pageSize),
		TotalItems:  total,
		HasNext:     hasNext,
		NextCursor:  nextCursor,
	}
}

func PaginationFromCursor(page, count, total, pageSize int, nextCursor string) common.PaginationInfo {
	if page <= 0 {
		page = 1
	}
	hasNext := nextCursor != "" && count > 0
	return common.PaginationInfo{
		CurrentPage: page,
		TotalPages:  TotalPages(total, pageSize),
		TotalItems:  total,
		HasNext:     hasNext,
		NextCursor:  nextCursor,
	}
}
