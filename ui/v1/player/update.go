package player

import (
	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/lazyspotify/core/ticker"
)

func (m *Model) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *Model) NextFrame(playing bool) {
	if playing {
		m.cassette.NextFrame()
	}
}

func (m *Model) NextButtonFrame() {
	for idx := range m.controls {
		if m.controls[idx].pressed {
			m.controls[idx].pressed = false
		}
	}
}

func (m *Model) HandleButtonPress(kind ButtonKind) tea.Cmd {
	if int(kind) < 0 || int(kind) >= len(m.controls) {
		return nil
	}
	m.controls[kind].pressed = true
	return ticker.DoTickClick()
}

func (m *Model) UpdateStatus(status Status) {
	s := &m.cassette.playerStatus
	s.Online = status.PlayerReady
	s.Playing = status.Playing
	s.CurrentMs = status.Position
	s.DurationMs = status.Duration
	s.Volume = status.Volume
	s.VolumeMax = status.MaxVolume
}

func (m *Model) ShowVolume() tea.Cmd {
	m.cassette.playerStatus.ShowVolume = true
	return ticker.DoTickVolume()
}

func (m *Model) HideVolume() {
	m.cassette.playerStatus.ShowVolume = false
}
