package viewbuilder

type ViewBuilder struct {
	view string

	lines int
}

func New() *ViewBuilder {
	return &ViewBuilder{}
}

func (vb *ViewBuilder) Add(s string) {
	for _, c := range s {
		if c == '\n' {
			vb.lines += 1
		}
	}

	vb.view += s
}

func (vb *ViewBuilder) AddLine() {
	vb.lines += 1
	vb.view += "\n"
}

func (vb *ViewBuilder) AddLines(n int) {
	if n < 0 {
		return
	}

	vb.lines += n
	for _ = range n {
		vb.view += "\n"
	}
}

func (vb *ViewBuilder) Lines() int {
	return vb.lines
}

func (vb *ViewBuilder) String() string {
	return vb.view
}
