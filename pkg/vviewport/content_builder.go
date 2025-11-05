package vviewport

import "strings"

type ContentBuilder struct {

	// dimensions of the virtual viewport
	yOffset int
	height  int

	view  []string
	lines int
}

// conditionally render the string if it exists within the virtual viewport
func (vb *ContentBuilder) AddCond(estLines int, s func() string) {

	// check if the estimated string overlaps the viewport
	if vb.lines > (vb.yOffset-estLines) && vb.lines < (vb.yOffset+vb.height+estLines) {
		vb.Add(s())
	}

}

func (vb *ContentBuilder) Add(s string) {
	a := s
	for {
		var (
			b  string
			ok bool
		)

		b, a, ok = strings.Cut(a, "\n")

		vb.lines += 1

		if vb.inView() {
			vb.view = append(vb.view, b)
		}

		// exit once there are no more line breaks
		if !ok {
			return
		}
	}
}

func (vb *ContentBuilder) inView() bool {
	return vb.lines > vb.yOffset && vb.lines < vb.height+vb.yOffset

}

func (vb *ContentBuilder) AddLine() {
	vb.AddLines(1)
}

func (vb *ContentBuilder) AddLines(n int) {
	if n < 0 {
		return
	}

	for range n {
		vb.lines += 1

		if vb.inView() {
			vb.view = append(vb.view, "")
		}
	}
}

func (vb *ContentBuilder) Lines() int {
	return vb.lines
}

func (vb *ContentBuilder) Content() []string {
	return vb.view
}

func (vb *ContentBuilder) String() string {
	return strings.Join(vb.view, "\n")
}
