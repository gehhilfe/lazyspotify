package mediacenter

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m *Model) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.keys.TogglePanel):
			m.mediaListOpen = !m.mediaListOpen
			if !m.mediaListOpen {
				m.mediaPanel.CloseInfo()
			}
			return nil
		case key.Matches(msg, m.keys.ZenMode):
			m.zenMode = !m.zenMode
		}
		if !m.mediaListOpen {
			return nil
		}
	}
	return m.mediaPanel.Update(msg)
}
