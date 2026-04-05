package v1

import (
	"strings"

	"charm.land/lipgloss/v2"
)

var (
	shellStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	labelStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("230")).Background(lipgloss.Color("24")).Padding(0, 1)
	metaStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	reelStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	trimStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("246"))
	windowStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("229"))
	footStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	screwStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	shineStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
)

type Cassette struct {
	spokeLeft  Spoke
	spokeRight Spoke
}

func NewCassette() Cassette {
	return Cassette{
		spokeLeft:  NewSpoke(),
		spokeRight: NewSpoke(),
	}
}

func (c *Cassette) NextFrame() {
	c.spokeLeft.NextFrame()
	c.spokeRight.NextFrame()
}

func (c *Cassette) View() string {
	compositor := lipgloss.NewCompositor(c.cassetteLayers()...)
	return compositor.Render()
}

func (c *Cassette) cassetteLayers() []*lipgloss.Layer {
	leftReel := c.spokeLeft.View()
	rightReel := c.spokeRight.View()
	label := cassetteLabel()
	meta := cassetteMeta()
	windowTrim := cassetteWindowTrim()
	tapeWindow := cassetteTapeWindow()
	foot := cassetteDeckFoot()
	shine := cassetteShine()

	leftReelW, leftReelH := lipgloss.Width(leftReel), lipgloss.Height(leftReel)
	metaW := lipgloss.Width(meta)
	windowTrimW, windowTrimH := lipgloss.Width(windowTrim), lipgloss.Height(windowTrim)
	tapeWindowW, tapeWindowH := lipgloss.Width(tapeWindow), lipgloss.Height(tapeWindow)
	labelW := lipgloss.Width(label)
	footW := lipgloss.Width(foot)
	shineW := lipgloss.Width(shine)

	sidePad := 3
	centerGap := maxInt(windowTrimW+4, tapeWindowW+6, 12)
	innerW := sidePad*2 + leftReelW + centerGap + leftReelW
	innerW = maxInt(innerW, labelW+6, metaW+6, footW*2+12)

	reelYInner := 3
	innerH := maxInt(reelYInner+leftReelH+3, 11)

	shell := cassetteShell(innerW, innerH)
	shellW := lipgloss.Width(shell)
	shellH := lipgloss.Height(shell)

	leftReelX := 1 + sidePad
	leftReelY := 1 + reelYInner
	rightReelX := shellW - 1 - sidePad - leftReelW
	rightReelY := leftReelY

	gapStart := leftReelX + leftReelW
	gapWidth := rightReelX - gapStart

	labelX := 1 + (innerW-labelW)/2
	labelY := 2
	metaX := 1 + (innerW-metaW)/2
	metaY := labelY + 1
	windowTrimX := gapStart + (gapWidth-windowTrimW)/2
	windowTrimY := leftReelY + (leftReelH-windowTrimH)/2 - 1
	tapeWindowX := gapStart + (gapWidth-tapeWindowW)/2
	tapeWindowY := leftReelY + (leftReelH-tapeWindowH)/2 + 1

	leftFootX := 1 + sidePad
	leftFootY := shellH - 4
	rightFootX := shellW - 1 - sidePad - footW
	rightFootY := leftFootY

	screwOffsetX := 2
	topScrewY := 1
	bottomScrewY := shellH - 3
	shineX := shellW - 1 - shineW - 2
	shineY := 1

	return []*lipgloss.Layer{
		lipgloss.NewLayer(shellStyle.Render(shell)).ID("shell"),
		lipgloss.NewLayer(labelStyle.Render(label)).X(labelX).Y(labelY).ID("label"),
		lipgloss.NewLayer(metaStyle.Render(meta)).X(metaX).Y(metaY).ID("meta"),
		lipgloss.NewLayer(reelStyle.Render(leftReel)).X(leftReelX).Y(leftReelY).ID("left-reel"),
		lipgloss.NewLayer(reelStyle.Render(rightReel)).X(rightReelX).Y(rightReelY).ID("right-reel"),
		lipgloss.NewLayer(trimStyle.Render(windowTrim)).X(windowTrimX).Y(windowTrimY).ID("window-trim"),
		lipgloss.NewLayer(windowStyle.Render(tapeWindow)).X(tapeWindowX).Y(tapeWindowY).ID("tape-window"),
		lipgloss.NewLayer(footStyle.Render(foot)).X(leftFootX).Y(leftFootY).ID("left-foot"),
		lipgloss.NewLayer(footStyle.Render(foot)).X(rightFootX).Y(rightFootY).ID("right-foot"),
		lipgloss.NewLayer(screwStyle.Render(cassetteScrew())).X(screwOffsetX).Y(topScrewY).ID("screw-tl"),
		lipgloss.NewLayer(screwStyle.Render(cassetteScrew())).X(shellW - screwOffsetX - 1).Y(topScrewY).ID("screw-tr"),
		lipgloss.NewLayer(screwStyle.Render(cassetteScrew())).X(screwOffsetX).Y(bottomScrewY).ID("screw-bl"),
		lipgloss.NewLayer(screwStyle.Render(cassetteScrew())).X(shellW - screwOffsetX - 1).Y(bottomScrewY).ID("screw-br"),
		lipgloss.NewLayer(shineStyle.Render(shine)).X(shineX).Y(shineY).ID("shine"),
	}
}

func cassetteShell(innerW, innerH int) string {
	lines := make([]string, 0, innerH+2)
	lines = append(lines, "."+strings.Repeat("-", innerW)+".")
	for i := 0; i < innerH; i++ {
		fill := strings.Repeat(" ", innerW)
		if i == innerH-1 {
			fill = strings.Repeat("_", innerW)
		}
		lines = append(lines, "|"+fill+"|")
	}
	lines = append(lines, "'"+strings.Repeat("-", innerW)+"'")
	return strings.Join(lines, "\n")
}

func cassetteLabel() string {
	return "LAZYSPOTIFY  C-60"
}

func cassetteMeta() string {
	return "TYPE I  |  STEREO  |  SIDE A"
}

func cassetteWindowTrim() string {
	lines := []string{
		".___          ___.",
		"|   |        |   |",
		"'---'        '---'",
	}
	return strings.Join(lines, "\n")
}

func cassetteTapeWindow() string {
	lines := []string{
		".------.",
		"|==  ==|",
		"'------'",
	}
	return strings.Join(lines, "\n")
}

func cassetteDeckFoot() string {
	return "[====]"
}

func cassetteShine() string {
	return "///"
}

func cassetteScrew() string {
	return "*"
}

func maxInt(values ...int) int {
	if len(values) == 0 {
		return 0
	}
	m := values[0]
	for _, v := range values[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
