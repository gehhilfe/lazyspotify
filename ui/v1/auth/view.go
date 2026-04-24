package auth

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) View() tea.View {
	if m.kind == KindNavidrome {
		return m.viewNavidrome()
	}
	return m.viewSpotify()
}

func (m *Model) viewSpotify() tea.View {
	if m.err != nil {
		return tea.NewView(fmt.Sprintf("Authentication failed: %v\nExiting...", m.err))
	}
	if m.auth.AuthServer.Started.Load() {
		head := lipgloss.NewStyle().Width(m.width).Foreground(lipgloss.BrightYellow).MarginBottom(1).Render("Authenticating with Spotify")
		msg := lipgloss.NewStyle().Width(m.width).MarginBottom(1).Render("Please open this link in your browser")
		styledURL := lipgloss.NewStyle().Width(m.width).Foreground(lipgloss.BrightBlue).Render(m.auth.GetAuthURL())
		hintText := "press c to copy"
		hintColor := lipgloss.BrightBlack
		if m.copied {
			hintText = "✓ copied to clipboard"
			hintColor = lipgloss.BrightGreen
		}
		hint := lipgloss.NewStyle().Foreground(hintColor).MarginTop(3).Render(hintText)
		return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, head, msg, styledURL, hint))
	}
	return tea.NewView("Authenticating with Spotify")
}

func (m *Model) viewNavidrome() tea.View {
	head := lipgloss.NewStyle().Width(m.width).Foreground(lipgloss.BrightYellow).MarginBottom(1).Render("Navidrome login")
	server := lipgloss.NewStyle().Width(m.width).Render(fmt.Sprintf("Server: %s", m.ndAuth.ServerURL()))
	user := lipgloss.NewStyle().Width(m.width).MarginBottom(1).Render(fmt.Sprintf("User:   %s", m.ndAuth.Username()))

	var status string
	switch m.authState {
	case Authenticating:
		status = lipgloss.NewStyle().Foreground(lipgloss.BrightBlue).Render("Validating…")
	case Authenticated:
		status = lipgloss.NewStyle().Foreground(lipgloss.BrightGreen).Render("Authenticated")
	default:
		status = lipgloss.NewStyle().Foreground(lipgloss.BrightBlack).Render("Enter your password and press Enter")
	}

	input := m.pwInput.View()

	parts := []string{head, server, user, input, status}
	if m.err != nil {
		errLine := lipgloss.NewStyle().Foreground(lipgloss.BrightRed).MarginTop(1).Render(fmt.Sprintf("Error: %v", m.err))
		parts = append(parts, errLine)
	}
	return tea.NewView(lipgloss.JoinVertical(lipgloss.Left, parts...))
}
