package viewport2

import (
	"math"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

// New returns a new model with the given width and height as well as default
// key mappings.
func New(width, height int) (m Model) {
	m.Width = width
	m.Height = height
	m.setInitialValues()
	return m
}

// Model is the Bubble Tea model for this viewport element.
type Model struct {
	Width  int
	Height int
	KeyMap KeyMap

	// Whether or not to respond to the mouse. The mouse must be enabled in
	// Bubble Tea for this to work. For details, see the Bubble Tea docs.
	MouseWheelEnabled bool

	// The number of lines the mouse wheel will scroll. By default, this is 3.
	MouseWheelDelta int

	// YOffset is the vertical scroll position.
	YOffset int

	// xOffset is the horizontal scroll position.
	xOffset int

	// horizontalStep is the number of columns we move left or right during a
	// default horizontal scroll.
	horizontalStep int

	// YPosition is the position of the viewport in relation to the terminal
	// window. It's used in high performance rendering only.
	YPosition int

	// Style applies a lipgloss style to the viewport. Realistically, it's most
	// useful for setting borders, margins and padding.
	Style lipgloss.Style

	initialized bool

	virtualContents *VirtualContents

	longestLineWidth int
}

func (m *Model) setInitialValues() {
	m.KeyMap = DefaultKeyMap()
	m.MouseWheelEnabled = true
	m.MouseWheelDelta = 3
	m.initialized = true
}

// Init exists to satisfy the tea.Model interface for composability purposes.
func (m Model) Init() tea.Cmd {
	return nil
}

// AtTop returns whether or not the viewport is at the very top position.
func (m Model) AtTop() bool {
	return m.YOffset <= 0
}

// AtBottom returns whether or not the viewport is at or past the very bottom
// position.
func (m Model) AtBottom() bool {
	return m.YOffset >= m.maxYOffset()
}

// PastBottom returns whether or not the viewport is scrolled beyond the last
// line. This can happen when adjusting the viewport height.
func (m Model) PastBottom() bool {
	return m.YOffset > m.maxYOffset()
}

// ScrollPercent returns the amount scrolled as a float between 0 and 1.
func (m Model) ScrollPercent() float64 {
	if m.Height >= m.virtualContents.Pos() {
		return 1.0
	}
	y := float64(m.YOffset)
	h := float64(m.Height)
	t := float64(m.virtualContents.Pos())
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// HorizontalScrollPercent returns the amount horizontally scrolled as a float
// between 0 and 1.
func (m Model) HorizontalScrollPercent() float64 {
	if m.xOffset >= m.longestLineWidth-m.Width {
		return 1.0
	}
	y := float64(m.xOffset)
	h := float64(m.Width)
	t := float64(m.longestLineWidth)
	v := y / (t - h)
	return math.Max(0.0, math.Min(1.0, v))
}

// SetContent set the pager's text content.
func (m *Model) SetContent(vc *VirtualContents) {
	m.virtualContents = vc
	m.longestLineWidth = findLongestLineWidth(vc.Lines())

	// TODO: come back to this
	if m.YOffset > m.virtualContents.Pos()-1 {
		m.GotoBottom()
	}
}

// maxYOffset returns the maximum possible value of the y-offset based on the
// viewport's content and set height.
func (m Model) maxYOffset() int {
	return max(0, m.virtualContents.Pos()-m.Height+m.Style.GetVerticalFrameSize())
}

// visibleLines returns the lines that should currently be visible in the
// viewport.
func (m Model) visibleLines() (lines []string) {
	return m.virtualContents.Lines()
}

// scrollArea returns the scrollable boundaries for high performance rendering.
//
// Deprecated: high performance rendering is deprecated in Bubble Tea.
func (m Model) scrollArea() (top, bottom int) {
	top = max(0, m.YPosition)
	bottom = max(top, top+m.Height)
	if top > 0 && bottom > top {
		bottom--
	}
	return top, bottom
}

// SetYOffset sets the Y offset.
func (m *Model) SetYOffset(n int) {
	m.YOffset = clamp(n, 0, m.maxYOffset())
}

// PageDown moves the view down by the number of lines in the viewport.
func (m *Model) PageDown() {
	if m.AtBottom() {
		return
	}

	m.ScrollDown(m.Height)
}

// PageUp moves the view up by one height of the viewport.
func (m *Model) PageUp() {
	if m.AtTop() {
		return
	}

	m.ScrollUp(m.Height)
}

// HalfPageDown moves the view down by half the height of the viewport.
func (m *Model) HalfPageDown() {
	if m.AtBottom() {
		return
	}

	m.ScrollDown(m.Height / 2) //nolint:mnd
}

// HalfPageUp moves the view up by half the height of the viewport.
func (m *Model) HalfPageUp() {
	if m.AtTop() {
		return
	}

	m.ScrollUp(m.Height / 2) //nolint:mnd
}

// ScrollDown moves the view down by the given number of lines.
func (m *Model) ScrollDown(n int) {
	if m.AtBottom() || n == 0 || m.virtualContents.Pos() == 0 {
		return
	}

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we actually have left before we reach
	// the bottom.
	m.SetYOffset(m.YOffset + n)
}

// ScrollUp moves the view down by the given number of lines. Returns the new
// lines to show.
func (m *Model) ScrollUp(n int) {
	if m.AtTop() || n == 0 || m.virtualContents.Pos() == 0 {
		return
	}

	// Make sure the number of lines by which we're going to scroll isn't
	// greater than the number of lines we are from the top.
	m.SetYOffset(m.YOffset - n)
}

// SetHorizontalStep sets the default amount of columns to scroll left or right
// with the default viewport key map.
//
// If set to 0 or less, horizontal scrolling is disabled.
//
// On v1, horizontal scrolling is disabled by default.
func (m *Model) SetHorizontalStep(n int) {
	m.horizontalStep = max(n, 0)
}

// SetXOffset sets the X offset.
func (m *Model) SetXOffset(n int) {
	m.xOffset = clamp(n, 0, m.longestLineWidth-m.Width)
}

// ScrollLeft moves the viewport to the left by the given number of columns.
func (m *Model) ScrollLeft(n int) {
	m.SetXOffset(m.xOffset - n)
}

// ScrollRight moves viewport to the right by the given number of columns.
func (m *Model) ScrollRight(n int) {
	m.SetXOffset(m.xOffset + n)
}

// GotoTop sets the viewport to the top position.
func (m *Model) GotoTop() (lines []string) {
	if m.AtTop() {
		return nil
	}

	m.SetYOffset(0)
	return m.visibleLines()
}

// GotoBottom sets the viewport to the bottom position.
func (m *Model) GotoBottom() (lines []string) {
	m.SetYOffset(m.maxYOffset())
	return m.visibleLines()
}

// Update handles standard message-based viewport updates.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m, cmd = m.updateAsModel(msg)
	return m, cmd
}

// Author's note: this method has been broken out to make it easier to
// potentially transition Update to satisfy tea.Model.
func (m Model) updateAsModel(msg tea.Msg) (Model, tea.Cmd) {
	if !m.initialized {
		m.setInitialValues()
	}

	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.KeyMap.PageDown):
			m.PageDown()

		case key.Matches(msg, m.KeyMap.PageUp):
			m.PageUp()

		case key.Matches(msg, m.KeyMap.HalfPageDown):
			m.HalfPageDown()

		case key.Matches(msg, m.KeyMap.HalfPageUp):
			m.HalfPageUp()

		case key.Matches(msg, m.KeyMap.Down):
			m.ScrollDown(1)

		case key.Matches(msg, m.KeyMap.Up):
			m.ScrollUp(1)

		case key.Matches(msg, m.KeyMap.Left):
			m.ScrollLeft(m.horizontalStep)

		case key.Matches(msg, m.KeyMap.Right):
			m.ScrollRight(m.horizontalStep)
		}

	case tea.MouseMsg:
		if !m.MouseWheelEnabled || msg.Action != tea.MouseActionPress {
			break
		}
		switch msg.Button { //nolint:exhaustive
		case tea.MouseButtonWheelUp:
			if msg.Shift {
				// Note that not every terminal emulator sends the shift event for mouse actions by default (looking at you Konsole)
				m.ScrollLeft(m.horizontalStep)
			} else {
				m.ScrollUp(m.MouseWheelDelta)
			}

		case tea.MouseButtonWheelDown:
			if msg.Shift {
				m.ScrollRight(m.horizontalStep)
			} else {
				m.ScrollDown(m.MouseWheelDelta)
			}
		// Note that not every terminal emulator sends the horizontal wheel events by default (looking at you Konsole)
		case tea.MouseButtonWheelLeft:
			m.ScrollLeft(m.horizontalStep)
		case tea.MouseButtonWheelRight:
			m.ScrollRight(m.horizontalStep)
		}
	}

	return m, cmd
}

// View renders the viewport into a string.
func (m Model) View() string {
	w, h := m.Width, m.Height
	if sw := m.Style.GetWidth(); sw != 0 {
		w = min(w, sw)
	}
	if sh := m.Style.GetHeight(); sh != 0 {
		h = min(h, sh)
	}
	contentWidth := w - m.Style.GetHorizontalFrameSize()
	contentHeight := h - m.Style.GetVerticalFrameSize()
	contents := lipgloss.NewStyle().
		Width(contentWidth).      // pad to width.
		Height(contentHeight).    // pad to height.
		MaxHeight(contentHeight). // truncate height if taller.
		MaxWidth(contentWidth).   // truncate width if wider.
		Render(strings.Join(m.visibleLines(), "\n"))
	return m.Style.
		UnsetWidth().UnsetHeight(). // Style size already applied in contents.
		Render(contents)
}

func clamp(v, low, high int) int {
	if high < low {
		low, high = high, low
	}
	return min(high, max(low, v))
}

func findLongestLineWidth(lines []string) int {
	w := 0
	for _, l := range lines {
		if ww := ansi.StringWidth(l); ww > w {
			w = ww
		}
	}
	return w
}
