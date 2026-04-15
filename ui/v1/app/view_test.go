package app

import (
	"strings"
	"testing"
)

func TestViewShowsSmallTerminalMessageWhenViewportIsTooSmall(t *testing.T) {
	model := NewModel()
	model.setSize(20, 6)

	view := model.View()

	if !strings.Contains(view.Content, "terminal size too") || !strings.Contains(view.Content, "small") {
		t.Fatalf("view = %q, want small terminal warning", view.Content)
	}
}

func TestViewRendersMainUIWhenViewportFits(t *testing.T) {
	model := NewModel()
	model.setSize(120, 40)

	view := model.View()

	if strings.Contains(view.Content, "terminal size too small") {
		t.Fatalf("view = %q, did not expect small terminal warning", view.Content)
	}
}
