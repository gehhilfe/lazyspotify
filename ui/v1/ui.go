package v1

import (
	"os"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
)

func newModel() Model {
	return Model{
		mediaCenter: NewMediaCenter(),
	}
}

func (m *Model) Init() tea.Cmd {
	cmd := func() tea.Msg {
		err := m.start()
		if err != nil && !m.authModel.needsAuth {
			return tea.Msg(err)
		}
		if m.authModel.needsAuth {
			return tea.Msg(m.authModel.needsAuth)
		}
		return startupCompleteMsg{}
	}

	return tea.Batch(cmd, DoTickSpokes())
}

func (m *Model) View() tea.View {
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	if m.authModel != nil && m.authModel.needsAuth {
		return m.authModel.View()
	}
	mediaCenter := m.mediaCenter
	v := mediaCenter.View()
	return tea.NewView(v + "\n" + helpStyle.Render("Press q to quit"))
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	centerCmd := m.mediaCenter.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		return m, centerCmd
	case NextSpokeFrameMsg:
		m.NextFrame()
		return m, tea.Batch(DoTickSpokes(), centerCmd)
	case NextButtonFrameMsg:
		m.NextButtonFrame()
		return m, centerCmd
	case MediaRequest:
		startCmd := m.mediaCenter.StartLoading()
		fetchCmd := m.HandleMediaRequest(msg)
		return m, tea.Batch(startCmd, fetchCmd, centerCmd)
	case startupCompleteMsg:
		requestCmd := tea.Cmd(func() tea.Msg {
			return MediaRequest{kind: GetUserLibrary, offset: 0}
		})
		return m, tea.Batch(m.waitForPlayerReady(), requestCmd, centerCmd)
	case playerReadyMsg:
		m.playDailyMix()
		return m, centerCmd
	case playerReadyErrMsg:
		logger.Log.Error().Err(msg.err).Msg("failed to wait for player to be ready")
		return m, centerCmd
	case mediaLoadedMsg:
		setContentCmd := m.mediaCenter.SetContent(msg.entities, msg.kind)
		return m, tea.Batch(setContentCmd, centerCmd)
	case mediaLoadErrMsg:
		logger.Log.Error().Err(msg.err).Msg("failed to get user library")
		m.mediaCenter.visibleList.list.StopSpinner()
		return m, tea.Batch(m.mediaCenter.visibleList.list.NewStatusMessage("Failed to load library"), centerCmd)
	}
	if m.authModel != nil && m.authModel.needsAuth {
		newM, cmd := m.authModel.Update(msg)
		m.authModel = newM.(*AuthModel)
		return m, tea.Batch(cmd, centerCmd)
	}
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch msg.String() {
		case " ", "p":
			m.playing = !m.playing
			if m.playing {
				m.HandleButtonPress(PlayButton)
			} else {
				m.HandleButtonPress(PauseButton)
			}
			m.playPause()
			return m, tea.Batch(cmd, DoTickButtonPress(), centerCmd)
		case "right", "l", "ctrl+f", "]":
			m.HandleButtonPress(SeekForwardButton)
			m.seekForward()
			return m, tea.Batch(cmd, DoTickButtonPress(), centerCmd)
		case "left", "h", "ctrl+b", "[":
			m.HandleButtonPress(SeekBackwardButton)
			m.seekBackward()
			return m, tea.Batch(cmd, DoTickButtonPress(), centerCmd)
		case "n", "ctrl+s":
			m.HandleButtonPress(NextButton)
			m.next()
			return m, tea.Batch(cmd, DoTickButtonPress(), centerCmd)
		case "N", "ctrl+r":
			m.HandleButtonPress(PreviousButton)
			m.previous()
			return m, tea.Batch(cmd, DoTickButtonPress(), centerCmd)
		case "j", "down", "ctrl+p":
			m.decrementVolume()
			return m, tea.Batch(cmd, centerCmd)
		case "k", "up", "ctrl+n":
			m.incrementVolume()
			return m, tea.Batch(cmd, centerCmd)
		}
	}
	return m, centerCmd
}

func RunTui() {
	model := newModel()
	_, err := tea.NewProgram(&model).Run()
	model.shutdown()
	if err != nil {
		logger.Log.Error().Err(err).Msg("failed to run program")
		os.Exit(1)
	}
}
