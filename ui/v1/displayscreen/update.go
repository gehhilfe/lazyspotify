package displayscreen

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	ansi "github.com/charmbracelet/x/ansi"
	"github.com/dubeyKartikay/lazyspotify/core/ticker"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
)

func (m *Model) SetDisplayFromSong(song common.SongInfo) {
	if song.Title == "" {
		return
	}
	separator := " • "
	m.display = m.styles.primary.Render(song.Title) +
		m.styles.accent.Render(separator) +
		m.styles.muted.Render(song.Artist) +
		m.styles.accent.Render(separator) +
		m.styles.muted.Render(song.Album)
}

func (m *Model) Update(tea.Msg) tea.Cmd {
	return nil
}

func (m *Model) NextFrame() tea.Cmd {
	if len(m.display) > 0 {
		m.scrollOffset = (m.scrollOffset + 1) % len(m.display)
	}
	return ticker.DoTick()
}

func (m *Model) scrollText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipglossWidth(text) <= width {
		return text
	}

	const gap = "   "
	base := text + gap
	track := base + base
	if len(base) == 0 {
		return strings.Repeat(" ", width)
	}

	start := m.scrollOffset
	end := min(len(track), start+width)
	return ansi.Cut(track, start, end)
}

func lipglossWidth(text string) int {
	return len([]rune(ansi.Strip(text)))
}
