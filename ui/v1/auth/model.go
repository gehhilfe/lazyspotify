package auth

import (
	"charm.land/bubbles/v2/textinput"
	"github.com/dubeyKartikay/lazyspotify/core/auth"
	"github.com/dubeyKartikay/lazyspotify/core/backend/navidrome"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
)

type State int

const (
	NeedsAuth State = iota
	Authenticating
	Authenticated
)

type Kind string

const (
	KindSpotify   Kind = "spotify"
	KindNavidrome Kind = "navidrome"
)

type Model struct {
	kind            Kind
	authState       State
	auth            *auth.Authenticator
	ndAuth          *navidrome.Authenticator
	pwInput         textinput.Model
	authFlowUpdates chan string
	err             error
	width           int
	height          int
	copied          bool
}

func NewModel() *Model {
	kind := resolveKind()
	m := &Model{
		kind:            kind,
		authState:       Authenticated,
		authFlowUpdates: make(chan string),
	}
	switch kind {
	case KindNavidrome:
		m.ndAuth = navidrome.NewAuthenticator()
		ti := textinput.New()
		ti.Placeholder = "password"
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
		ti.Prompt = "» "
		m.pwInput = ti
	default:
		m.auth = auth.New()
	}
	return m
}

func resolveKind() Kind {
	backend := utils.GetConfig().Backend
	if backend == utils.BackendNavidrome {
		return KindNavidrome
	}
	return KindSpotify
}

func (m *Model) Kind() Kind {
	return m.kind
}

func (m *Model) State() State {
	return m.authState
}

func (m *Model) SetState(state State) {
	m.authState = state
}

func (m *Model) Authenticator() *auth.Authenticator {
	return m.auth
}

func (m *Model) NavidromeAuth() *navidrome.Authenticator {
	return m.ndAuth
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.pwInput.SetWidth(width)
}
