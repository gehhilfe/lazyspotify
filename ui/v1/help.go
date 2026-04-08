package v1

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
)

type appKeyMap struct {
	ToggleHelp   key.Binding
	Quit         key.Binding
	CycleLibrary key.Binding
	Select       key.Binding
	Back         key.Binding
	NextPage     key.Binding
	PrevPage     key.Binding
	TogglePanel  key.Binding
	PlayPause    key.Binding
	SeekForward  key.Binding
	SeekBackward key.Binding
	NextTrack    key.Binding
	PrevTrack    key.Binding
	VolumeDown   key.Binding
	VolumeUp     key.Binding
}

func newAppKeyMap() appKeyMap {
	return appKeyMap{
		ToggleHelp: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "toggle help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c", "esc"),
			key.WithHelp("q/esc", "quit"),
		),
		CycleLibrary: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "cycle library"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select item"),
		),
		Back: key.NewBinding(
			key.WithKeys("backspace", "delete"),
			key.WithHelp("del", "back"),
		),
		NextPage: key.NewBinding(
			key.WithKeys("ctrl+d"),
			key.WithHelp("ctrl+d", "next page"),
		),
		PrevPage: key.NewBinding(
			key.WithKeys("ctrl+u"),
			key.WithHelp("ctrl+u", "prev page"),
		),
		TogglePanel: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "toggle panel"),
		),
		PlayPause: key.NewBinding(
			key.WithKeys(" ", "space", "p"),
			key.WithHelp("space/p", "play pause"),
		),
		SeekForward: key.NewBinding(
			key.WithKeys("right", "l", "ctrl+f", "]"),
			key.WithHelp("right/l", "seek +"),
		),
		SeekBackward: key.NewBinding(
			key.WithKeys("left", "h", "ctrl+b", "["),
			key.WithHelp("left/h", "seek -"),
		),
		NextTrack: key.NewBinding(
			key.WithKeys("n", "ctrl+s"),
			key.WithHelp("n", "next"),
		),
		PrevTrack: key.NewBinding(
			key.WithKeys("N", "ctrl+r"),
			key.WithHelp("N", "previous"),
		),
		VolumeDown: key.NewBinding(
			key.WithKeys("j", "ctrl+p"),
			key.WithHelp("j", "volume -"),
		),
		VolumeUp: key.NewBinding(
			key.WithKeys("k", "ctrl+n"),
			key.WithHelp("k", "volume +"),
		),
	}
}

func (k appKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.ToggleHelp}
}

func (k appKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ToggleHelp, k.Quit, k.CycleLibrary, k.Select},
		{k.Back, k.NextPage, k.PrevPage, k.TogglePanel, k.PlayPause},
		{k.SeekForward, k.SeekBackward, k.VolumeDown, k.VolumeUp},
		{k.NextTrack, k.PrevTrack},
	}
}

func newHelpModel() help.Model {
	h := help.New()
	h.ShowAll = false
	return h
}
