package navidrome

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"

	"github.com/dubeyKartikay/lazyspotify/core/utils"
	"github.com/zalando/go-keyring"
)

// Authenticator loads and stores Navidrome credentials. Server URL and
// username live in config; the password lives in the system keyring.
type Authenticator struct {
	cfg utils.AppConfig
}

func NewAuthenticator() *Authenticator {
	return &Authenticator{cfg: utils.GetConfig()}
}

func (a *Authenticator) ServerURL() string {
	return strings.TrimSpace(a.cfg.Navidrome.ServerURL)
}

func (a *Authenticator) Username() string {
	return strings.TrimSpace(a.cfg.Navidrome.Username)
}

func (a *Authenticator) Password() (string, error) {
	pw, err := keyring.Get(a.cfg.Navidrome.Keyring.Service, a.cfg.Navidrome.Keyring.Key)
	if err != nil {
		return "", wrapKeyringError(err)
	}
	return pw, nil
}

func (a *Authenticator) SetPassword(password string) error {
	if strings.TrimSpace(password) == "" {
		return fmt.Errorf("password is empty")
	}
	return wrapKeyringError(keyring.Set(a.cfg.Navidrome.Keyring.Service, a.cfg.Navidrome.Keyring.Key, password))
}

func (a *Authenticator) ClearPassword() error {
	return wrapKeyringError(keyring.Delete(a.cfg.Navidrome.Keyring.Service, a.cfg.Navidrome.Keyring.Key))
}

// Credentials resolves config + keyring into a Credentials struct ready to
// pass to NewClient. Returns ErrPasswordMissing if no password is stored.
func (a *Authenticator) Credentials() (Credentials, error) {
	if a.ServerURL() == "" {
		return Credentials{}, fmt.Errorf("navidrome server_url not configured")
	}
	if a.Username() == "" {
		return Credentials{}, fmt.Errorf("navidrome username not configured")
	}
	pw, err := a.Password()
	if err != nil {
		return Credentials{}, err
	}
	return Credentials{
		ServerURL:          a.ServerURL(),
		Username:           a.Username(),
		Password:           pw,
		InsecureSkipVerify: a.cfg.Navidrome.InsecureSkipVerify,
	}, nil
}

// Validate builds a client from the given password and pings the server.
func (a *Authenticator) Validate(ctx context.Context, password string) error {
	client, err := NewClient(Credentials{
		ServerURL:          a.ServerURL(),
		Username:           a.Username(),
		Password:           password,
		InsecureSkipVerify: a.cfg.Navidrome.InsecureSkipVerify,
	})
	if err != nil {
		return err
	}
	return client.Ping(ctx)
}

// IsPasswordMissing reports whether the given error is a keyring-miss.
func IsPasswordMissing(err error) bool {
	return errors.Is(err, keyring.ErrNotFound)
}

func wrapKeyringError(err error) error {
	if err == nil || errors.Is(err, keyring.ErrNotFound) {
		return err
	}
	if runtime.GOOS == "linux" {
		return fmt.Errorf(
			"system keyring unavailable: %w; lazyspotify requires a working Linux keyring and will not fall back to plaintext password storage",
			err,
		)
	}
	return fmt.Errorf("system keyring unavailable: %w", err)
}
