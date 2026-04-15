package mediacenter

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/dubeyKartikay/lazyspotify/ui/v1/common"
)

func TestViewHidesPlayerWhenPanelIsOpenAndViewportIsTooSmall(t *testing.T) {
	model := NewModel(common.NewAppKeyMap())

	model.Update(tea.KeyPressMsg(tea.Key{Text: "P", Code: 'P'}))

	view := model.View(40, 24)

	if strings.Contains(view, "LAZYSPOTIFY") {
		t.Fatalf("view = %q, did not expect player to render in narrow viewport", view)
	}
}

func TestViewKeepsPlayerWhenPanelIsOpenAndViewportFits(t *testing.T) {
	model := NewModel(common.NewAppKeyMap())

	model.Update(tea.KeyPressMsg(tea.Key{Text: "P", Code: 'P'}))

	view := model.View(120, 40)

	if !strings.Contains(view, "LAZYSPOTIFY") {
		t.Fatalf("view = %q, expected player to render in wide viewport", view)
	}
}

func TestViewShowsPlayerAgainAfterClosingLibrary(t *testing.T) {
	model := NewModel(common.NewAppKeyMap())

	model.Update(tea.KeyPressMsg(tea.Key{Text: "P", Code: 'P'}))
	model.CloseLibrary()

	view := model.View(120, 40)

	if !strings.Contains(view, "LAZYSPOTIFY") {
		t.Fatalf("view = %q, expected player to render after closing library", view)
	}
}

func TestViewDoesNotHidePlayerWhenViewportIsTooSmallAndPanelIsClosed(t *testing.T) {
	model := NewModel(common.NewAppKeyMap())

	view := model.View(1, 1)

	if !strings.Contains(view, "LAZYSPOTIFY") {
		t.Fatalf("view = %q, expected player to remain rendered when panel is closed", view)
	}
}
