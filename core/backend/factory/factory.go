package factory

import (
	"context"
	"errors"
	"fmt"

	"github.com/dubeyKartikay/lazyspotify/core/auth"
	"github.com/dubeyKartikay/lazyspotify/core/backend"
	"github.com/dubeyKartikay/lazyspotify/core/backend/navidrome"
	backendspotify "github.com/dubeyKartikay/lazyspotify/core/backend/spotify"
	coreplayer "github.com/dubeyKartikay/lazyspotify/core/player"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
	"github.com/dubeyKartikay/lazyspotify/spotify"
	"github.com/zalando/go-keyring"
)

type Kind string

const (
	KindSpotify   Kind = "spotify"
	KindNavidrome Kind = "navidrome"
)

type Bundle struct {
	Library backend.Library
	Player  backend.Player
}

// ErrNeedsAuth signals that the backend cannot be constructed because
// credentials are missing or invalid. The UI should prompt the user.
var ErrNeedsAuth = errors.New("backend needs authentication")

func New(ctx context.Context, authenticator *auth.Authenticator) (Bundle, error) {
	switch ResolveKind() {
	case KindNavidrome:
		return newNavidrome(ctx)
	default:
		return newSpotify(ctx, authenticator)
	}
}

func ResolveKind() Kind {
	cfg := utils.GetConfig()
	if cfg.Backend == utils.BackendNavidrome {
		return KindNavidrome
	}
	return KindSpotify
}

func newSpotify(ctx context.Context, authenticator *auth.Authenticator) (Bundle, error) {
	client, err := spotify.NewSpotifyClient(ctx, authenticator)
	if err != nil {
		if spotify.IsAuthError(err) {
			return Bundle{}, fmt.Errorf("%w: %v", ErrNeedsAuth, err)
		}
		return Bundle{}, err
	}
	userID, err := client.GetUserID(ctx)
	if err != nil {
		if spotify.IsAuthError(err) {
			return Bundle{}, fmt.Errorf("%w: %v", ErrNeedsAuth, err)
		}
		return Bundle{}, err
	}
	token, err := authenticator.GetAuthToken(ctx)
	if err != nil || token == nil {
		if err == nil {
			err = fmt.Errorf("missing spotify auth token")
		}
		return Bundle{}, fmt.Errorf("%w: %v", ErrNeedsAuth, err)
	}
	player, err := coreplayer.NewPlayer(ctx, userID, token.AccessToken)
	if err != nil {
		return Bundle{}, err
	}
	library := backendspotify.NewLibrary(client, player)
	return Bundle{Library: library, Player: player}, nil
}

func newNavidrome(ctx context.Context) (Bundle, error) {
	auth := navidrome.NewAuthenticator()
	creds, err := auth.Credentials()
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return Bundle{}, fmt.Errorf("%w: password not in keyring", ErrNeedsAuth)
		}
		return Bundle{}, err
	}
	client, err := navidrome.NewClient(creds)
	if err != nil {
		return Bundle{}, err
	}
	if err := client.Ping(ctx); err != nil {
		if navidrome.IsAuthError(err) {
			return Bundle{}, fmt.Errorf("%w: %v", ErrNeedsAuth, err)
		}
		return Bundle{}, err
	}
	library := navidrome.NewLibrary(client)
	player, err := navidrome.NewPlayer(client)
	if err != nil {
		return Bundle{}, err
	}
	return Bundle{Library: library, Player: player}, nil
}
