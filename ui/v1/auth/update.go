package auth

import (
	"context"
	"os/exec"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	coreauth "github.com/dubeyKartikay/lazyspotify/core/auth"
)

var keyMap = struct {
	CopyURL key.Binding
}{
	CopyURL: key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy url")),
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
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.authState == NeedsAuth {
		m.authState = Authenticating
		return m, tea.Batch(m.startAuthFlow, m.listenForAuthUpdates)
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keyMap.CopyURL) && m.auth.AuthServer.Started.Load() {
			url := m.auth.GetAuthURL()
			cmd := exec.Command("pbcopy")
			cmd.Stdin = strings.NewReader(url)
			m.copied = cmd.Run() == nil
		}
	case coreauth.AuthServerErr:
		m.err = msg.Err
		return m, tea.Quit
	case State:
		return m, m.listenForAuthUpdates
	}
	return m, nil
}
