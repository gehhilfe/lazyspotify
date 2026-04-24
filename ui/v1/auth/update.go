package auth

import (
	"context"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	coreauth "github.com/dubeyKartikay/lazyspotify/core/auth"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
)

type authQuitMsg struct{}
type navidromeValidationMsg struct {
	ok       bool
	password string
	err      error
}

var keyMap = struct {
	CopyURL key.Binding
	Submit  key.Binding
}{
	CopyURL: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy url")),
	Submit:  key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "submit")),
}

func (m *Model) startAuthFlow() tea.Msg {
	_, err := m.auth.ReAuthenticate(context.Background(), m.authFlowUpdates)
	if err != nil {
		return coreauth.AuthServerErr{Err: err}
	}
	return nil
}

func (m *Model) listenForAuthUpdates() tea.Msg {
	updates := <-m.authFlowUpdates
	if updates == "success" {
		m.authState = Authenticated
	}
	return m.authState
}

func (m *Model) Init() tea.Cmd {
	if m.kind == KindNavidrome {
		return m.pwInput.Focus()
	}
	return nil
}

func (m *Model) quitAfterError() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(2 * time.Second)
		return authQuitMsg{}
	}
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.kind == KindNavidrome {
		return m.updateNavidrome(msg)
	}
	return m.updateSpotify(msg)
}

func (m *Model) updateSpotify(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.authState == NeedsAuth {
		m.authState = Authenticating
		return m, tea.Batch(m.startAuthFlow, m.listenForAuthUpdates)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keyMap.CopyURL) && m.auth.AuthServer.Started.Load() {
			url := m.auth.GetAuthURL()
			m.copied = utils.CopyToClipboard(url) == nil
		}
	case coreauth.AuthServerErr:
		m.err = msg.Err
		return m, m.quitAfterError()
	case authQuitMsg:
		return m, tea.Quit
	case State:
		return m, m.listenForAuthUpdates
	}
	return m, nil
}

func (m *Model) updateNavidrome(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keyMap.Submit) && m.authState == NeedsAuth {
			pw := m.pwInput.Value()
			if pw == "" {
				m.err = errEmptyPassword{}
				return m, nil
			}
			m.err = nil
			m.authState = Authenticating
			return m, m.validateNavidrome(pw)
		}
	case navidromeValidationMsg:
		if msg.ok {
			if err := m.ndAuth.SetPassword(msg.password); err != nil {
				m.err = err
				m.authState = NeedsAuth
				return m, nil
			}
			m.authState = Authenticated
			return m, func() tea.Msg { return m.authState }
		}
		m.err = msg.err
		m.authState = NeedsAuth
		m.pwInput.Reset()
		return m, m.pwInput.Focus()
	case State:
		if msg == NeedsAuth && !m.pwInput.Focused() {
			return m, m.pwInput.Focus()
		}
		return m, nil
	}

	if m.authState == NeedsAuth {
		if !m.pwInput.Focused() {
			cmd := m.pwInput.Focus()
			m.pwInput, _ = m.pwInput.Update(msg)
			return m, cmd
		}
		var cmd tea.Cmd
		m.pwInput, cmd = m.pwInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m *Model) validateNavidrome(password string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := m.ndAuth.Validate(ctx, password); err != nil {
			logger.Log.Error().Err(err).Msg("navidrome validation failed")
			return navidromeValidationMsg{ok: false, err: err}
		}
		return navidromeValidationMsg{ok: true, password: password}
	}
}

type errEmptyPassword struct{}

func (errEmptyPassword) Error() string { return "password cannot be empty" }
