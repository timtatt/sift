package viewport2

import (
	"log/slog"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

type VirtualContents struct {
	lines   []string
	width   int
	height  int
	offsetY int

	pos        int
	widthStyle lipgloss.Style
}

func NewVirtualContents(width int, height int, offsetY int) *VirtualContents {
	return &VirtualContents{
		lines:   []string{},
		width:   width,
		height:  height,
		offsetY: offsetY,
		pos:     0,
	}
}

func wrapText(text string, width int) string {
	var output string
	lastBreak := 0
	for i, c := range text {
		lineLength := i - lastBreak
		if lineLength >= width {
			output += "\n"
			lastBreak = i
		}
		output += string(c)
	}

	return output
}

func (v *VirtualContents) AddCond(clen int, c func() string) {
	lineCount := lineCount(clen, v.width)

	if v.pos >= v.offsetY && v.pos <= v.offsetY+v.height {
		v.lines = append(v.lines, c())
	}

	v.pos += lineCount

}

func lineCount(len int, width int) int {
	return int(math.Ceil(float64(len) / float64(width)))
}

func (v *VirtualContents) Add(c string) {
	lineCount := lineCount(len(c), v.width)

	slog.Debug("adding line", "lineCount", lineCount)

	// only add the lines if we're within the window
	if v.pos >= v.offsetY && v.pos <= v.offsetY+v.height {
		v.lines = append(v.lines, c)
	}

	v.pos += lineCount
}

func (v *VirtualContents) AddLine() {
	v.Add("")
}

func (v *VirtualContents) Pos() int {
	return v.pos
}

func (v *VirtualContents) Lines() []string {
	return v.lines
}

func (v *VirtualContents) Render() string {
	return strings.Join(v.lines, "\n")
}
