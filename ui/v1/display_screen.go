package v1

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/dubeyKartikay/lazyspotify/core/logger"
	"github.com/dubeyKartikay/lazyspotify/core/ticker"
)

type displayScreen struct {
	songInfo       SongInfo
	width          int
	height         int
	defaultDisplay string
	scrollOffset   int
}

func newDisplayScreen() displayScreen {
	return displayScreen{
		defaultDisplay: "Lazyspotify: The cutest terminal music player, Lazyspotify: The cutest terminal music player, Lazyspotify: The cutest terminal music player",
	}
}

func (d *displayScreen) SetSongInfo(songInfo SongInfo) {
	d.songInfo = songInfo
}

func (d *displayScreen) View() string {
	songInfo := d.songInfo
	s := d.defaultDisplay
	if songInfo.title != "" {
		separator := " • "
		s = lipgloss.JoinHorizontal(lipgloss.Left, songInfo.title)
		s = lipgloss.JoinHorizontal(lipgloss.Left, s, separator, songInfo.artist)
		s = lipgloss.JoinHorizontal(lipgloss.Left, s, separator, songInfo.album)
	}

	contentWidth := max(0, d.width-2)
	if contentWidth > 0 {
		s = d.scrollText(s, contentWidth)
	}
	panel := lipgloss.NewStyle().Width(d.width).Height(d.height).BorderStyle(lipgloss.RoundedBorder()).Render(s)
	return panel
}

func (d *displayScreen) SetSize(width int, height int) {
	d.width = width
	d.height = height
}

func (d *displayScreen) NextFrame() tea.Cmd {
	d.scrollOffset++
	logger.Log.Debug().Int("scrollOffset", d.scrollOffset).Msg("scrolling")
	return ticker.DoTick()
}

func (d *displayScreen) scrollText(text string, width int) string {
	if width <= 0 {
		return ""
	}

	if lipgloss.Width(text) <= width {
		return text
	}

	const gap = "   "
	base := []rune(text + gap)
	track := append(base, base...)
	if len(base) == 0 {
		return strings.Repeat(" ", width)
	}

	start := d.scrollOffset % len(base)
	end := start + width
	if end > len(track) {
		end = len(track)
	}

	visible := string(track[start:end])
	if len([]rune(visible)) < width {
		visible += strings.Repeat(" ", width-len([]rune(visible)))
	}

	return visible
}
