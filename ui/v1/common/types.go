package common

type Entity struct {
	Name string
	Desc string
	ID   string
	Img  string
}

func NewEntity(name, desc, uri, img string) Entity {
	return Entity{
		Name: name,
		Desc: desc,
		ID:   uri,
		Img:  img,
	}
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
	Kind        MediaRequestKind
	Cursor      string
	Page        int
	EntityURI   string
	ContextURI  string
	ShowLoading bool
}

type PaginationInfo struct {
	CurrentPage int
	TotalPages  int
	TotalItems  int
	HasNext     bool
	NextCursor  string
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

type SongInfo struct {
	Title    string
	Artist   string
	Album    string
	Position int
	Duration int
}

type VolumeInfo struct {
	Volume int
	Max    int
}

func RequestKindForListKind(kind ListKind) MediaRequestKind {
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
	return MediaRequest{Kind: RequestKindForListKind(kind), Page: 1, ShowLoading: true}
}

func KindForRequestKind(kind MediaRequestKind) ListKind {
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

func ListTitle(kind ListKind) string {
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

func ListTitleAbbr(kind ListKind) string {
	switch kind {
	case Albums:
		return "AL"
	case Artists:
		return "AR"
	case Playlists:
		return "PL"
	case Tracks:
		return "TR"
	case Shows:
		return "SH"
	case Episodes:
		return "EP"
	case AudioBooks:
		return "AB"
	default:
		return "Media"
	}
}

func MapSlice[T any, U any](items []T, mapFn func(T) U) []U {
	mapped := make([]U, 0, len(items))
	for _, item := range items {
		mapped = append(mapped, mapFn(item))
	}
	return mapped
}
