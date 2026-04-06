package spotify

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/dubeyKartikay/lazyspotify/core/auth"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/zalando/go-keyring"
	"github.com/zmb3/spotify/v2"
	"golang.org/x/oauth2"
)

type SpotifyClient struct {
	client *spotify.Client
}

func NewSpotifyClient(ctx context.Context, auth *auth.Authenticator) (*SpotifyClient, error) {
	client, err := auth.GetClient(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("error getting spotify client")
		return nil, err
	}
	return &SpotifyClient{client: client}, nil
}

func (s *SpotifyClient) GetUserID(ctx context.Context) (string, error) {
	user, err := s.client.CurrentUser(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("error getting user id")
		return "", err
	}
	return user.ID, nil
}

func (s *SpotifyClient) GetFirstSavedTrack(ctx context.Context) (string, error) {
	tracks, err := s.GetSavedTracks(ctx, 0)
	if err != nil || tracks == nil || len(tracks.Tracks) == 0 {
		logger.Log.Error().Stack().Err(err).Msg("error getting daily mix")
		if err == nil {
			err = fmt.Errorf("no saved tracks found")
		}
		return "", err
	}
	logger.Log.Info().Any("playlists", tracks)
	return string(tracks.Tracks[0].URI), nil
}

func (s *SpotifyClient) GetUserPlaylists(ctx context.Context, offset int) (*spotify.SimplePlaylistPage, error) {
	list, err := s.client.CurrentUsersPlaylists(ctx, spotify.Offset(offset))
	if err != nil {
		logger.Log.Error().Err(err).Msg("error getting user playlists")
		return nil, err
	}
	return list, nil
}

func (s *SpotifyClient) GetSavedTracks(ctx context.Context, offset int) (*spotify.SavedTrackPage, error) {
	tracks, err := s.client.CurrentUsersTracks(ctx, spotify.Offset(offset))
	if err != nil {
		logger.Log.Error().Err(err).Msg("error getting saved tracks")
		return nil, err
	}
	return tracks, nil
}

func (s *SpotifyClient) GetSavedAlbums(ctx context.Context, offset int) (*spotify.SavedAlbumPage, error) {
	albums, err := s.client.CurrentUsersAlbums(ctx, spotify.Offset(offset))
	if err != nil {
		logger.Log.Error().Err(err).Msg("error getting saved albums")
		return nil, err
	}
	return albums, nil
}

func (s *SpotifyClient) GetFollowedArtists(ctx context.Context) (*spotify.FullArtistCursorPage, error) {
	artists, err := s.client.CurrentUsersFollowedArtists(ctx)
	if err != nil {
		logger.Log.Error().Err(err).Msg("error getting followed artists")
		return nil, err
	}
	return artists, nil
}

func IsAuthError(err error) bool {
	var spotifyErr spotify.Error
	if errors.Is(err, keyring.ErrNotFound) {
		return true
	}
	if errors.As(err, &spotifyErr) && spotifyErr.Status == http.StatusUnauthorized {
		return true
	}
	var retrieveErr *oauth2.RetrieveError
	return errors.As(err, &retrieveErr) && retrieveErr.ErrorCode == "invalid_grant"
}
