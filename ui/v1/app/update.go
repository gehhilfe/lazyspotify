package app

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/dubeyKartikay/lazyspotify/core/ticker"
	"github.com/dubeyKartikay/lazyspotify/core/utils"
	uiauth "github.com/dubeyKartikay/lazyspotify/ui/v1/auth"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/player"
)

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	centerCmd := m.mediaCenter.Update(msg)

	if cmd, handled := m.handleShellInput(msg, centerCmd); handled {
		return m, cmd
	}
	if cmd, handled := m.handleSystemMessages(msg, centerCmd); handled {
		return m, cmd
	}

	if m.authModel != nil && m.authModel.State() < uiauth.Authenticated {
		newModel, cmd := m.authModel.Update(msg)
		m.authModel = newModel.(*uiauth.Model)
		return m, cmd
	}

	if cmd, handled := m.handleTransportInput(msg, centerCmd); handled {
		return m, cmd
	}
	return m, centerCmd
}

func (m *Model) handleShellInput(msg tea.Msg, centerCmd tea.Cmd) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.ToggleHelp):
			m.help.ShowAll = !m.help.ShowAll
			return centerCmd, true
		case key.Matches(msg, m.keys.Quit):
			return tea.Quit, true
		}
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return centerCmd, true
	}
	return nil, false
}

func (m *Model) handleSystemMessages(msg tea.Msg, centerCmd tea.Cmd) (tea.Cmd, bool) {
	switch msg := msg.(type) {
	case uiauth.State:
		if msg == uiauth.Authenticated {
			logger.Log.Info().Msg("authenticated")
			return m.Init(), true
		}
	case ticker.TickFastMsg:
		m.advancePlayback(180)
		m.mediaCenter.TickPlayer(m.playing)
		return tea.Batch(ticker.DoTickFast(), centerCmd), true
	case ticker.TickMsg:
		return tea.Batch(m.mediaCenter.TickDisplay(), centerCmd), true
	case ticker.TickMsgVolume:
		m.mediaCenter.HideVolume()
		return centerCmd, true
	case ticker.TickClickMsg:
		m.mediaCenter.TickButtons()
		return centerCmd, true
	case common.MediaRequest:
		var startCmd tea.Cmd
		logger.Log.Info().Int("kind", int(msg.Kind)).Str("cursor", msg.Cursor).Int("page", msg.Page).Msg("requesting media")
		if msg.ShowLoading {
			startCmd = m.mediaCenter.StartLoading()
		}
		return tea.Batch(startCmd, m.handleMediaRequest(msg), centerCmd), true
	case startupCompleteMsg:
		requestCmd := tea.Cmd(func() tea.Msg {
			return common.MediaRequestForListKind(common.Playlists)
		})
		return tea.Batch(m.waitForPlayerReady(), m.waitForPlayerEvent(), requestCmd, centerCmd), true
	case playerReadyMsg:
		m.playerReady = true
		m.updatePlayerStatus()
		return centerCmd, true
	case playerReadyErrMsg:
		m.playerReady = false
		m.updatePlayerStatus()
		logger.Log.Error().Err(msg.err).Msg("failed to wait for player to be ready")
		return centerCmd, true
	case playerEventMsg:
		m.applyPlayerEvent(msg.event)
		m.updatePlayerStatus()
		return tea.Batch(m.waitForPlayerEvent(), centerCmd), true
	case playerEventsClosedMsg:
		logger.Log.Warn().Msg("player events stream closed")
		return centerCmd, true
	case mediaLoadedMsg:
		return tea.Batch(m.mediaCenter.SetContent(msg.entities, msg.kind, msg.pagination, msg.request), centerCmd), true
	case mediaLoadErrMsg:
		logger.Log.Error().Err(msg.err).Msg("failed to get user library")
		m.mediaCenter.StartLoading()
		return tea.Batch(m.mediaCenter.SetStatus("Failed to load library"), centerCmd), true
	case playTrackErrMsg:
		logger.Log.Error().Err(msg.err).Msg("failed to play track")
		return tea.Batch(m.mediaCenter.SetStatus("Failed to play track"), centerCmd), true
	case playTrackOkMsg:
		m.playing = true
		m.playerReady = true
		m.updatePlayerStatus()
		return tea.Batch(m.mediaCenter.SetStatus("Playing"), centerCmd), true
	}
	return nil, false
}

func (m *Model) handleTransportInput(msg tea.Msg, centerCmd tea.Cmd) (tea.Cmd, bool) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return nil, false
	}

	switch {
	case key.Matches(keyMsg, m.keys.PlayPause):
		m.playing = !m.playing
		button := player.PauseButton
		command := m.playPauseCmd()
		if m.playing {
			button = player.PlayButton
		}
		return tea.Batch(m.mediaCenter.PressButton(button), command, centerCmd), true
	case key.Matches(keyMsg, m.keys.SeekForward):
		return tea.Batch(m.mediaCenter.PressButton(player.SeekForwardButton), m.seekForwardCmd(), centerCmd), true
	case key.Matches(keyMsg, m.keys.SeekBackward):
		return tea.Batch(m.mediaCenter.PressButton(player.SeekBackwardButton), m.seekBackwardCmd(), centerCmd), true
	case key.Matches(keyMsg, m.keys.NextTrack):
		return tea.Batch(m.mediaCenter.PressButton(player.NextButton), m.nextCmd(), centerCmd), true
	case key.Matches(keyMsg, m.keys.PrevTrack):
		return tea.Batch(m.mediaCenter.PressButton(player.PreviousButton), m.previousCmd(), centerCmd), true
	case key.Matches(keyMsg, m.keys.VolumeDown):
		m.markVolumeOverlay()
		m.previewVolume(-utils.GetConfig().Librespot.VolumeStep)
		m.updatePlayerStatus()
		return tea.Batch(m.decrementVolumeCmd(), centerCmd), true
	case key.Matches(keyMsg, m.keys.VolumeUp):
		m.markVolumeOverlay()
		m.previewVolume(utils.GetConfig().Librespot.VolumeStep)
		m.updatePlayerStatus()
		return tea.Batch(m.incrementVolumeCmd(), centerCmd), true
	}

	return nil, false
}

func (m *Model) handleMediaRequest(request common.MediaRequest) tea.Cmd {
	if request.Page <= 0 {
		request.Page = 1
	}
	handler, ok := m.requestHandlers[request.Kind]
	if !ok {
		return nil
	}
	return handler(request)
}
