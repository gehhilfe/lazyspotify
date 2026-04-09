package player

type spoke struct {
	frames       []frame
	currentFrame int
}

type frame struct {
	view string
}

func newSpoke() spoke {
	top := " .---. "
	bottom := " '---' "

	frames := []frame{
		{view: top + "\n" + "/  |  \\" + "\n" + "|  o  |" + "\n" + "\\  |  /" + "\n" + bottom},
		{view: top + "\n" + "/   / \\" + "\n" + "|  o  |" + "\n" + "\\ /   /" + "\n" + bottom},
		{view: top + "\n" + "/     \\" + "\n" + "|--o--|" + "\n" + "\\     /" + "\n" + bottom},
		{view: top + "\n" + "/ \\   \\" + "\n" + "|  o  |" + "\n" + "\\   \\ /" + "\n" + bottom},
	}
	return spoke{frames: frames}
}

func (s *spoke) View() string {
	return s.frames[s.currentFrame].view
}

func (s *spoke) NextFrame() {
	s.currentFrame = (s.currentFrame + 1) % len(s.frames)
}
