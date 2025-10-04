package helpview

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// KeyMap is a map of keybindings used to generate help.
type KeyMap interface {
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

type WrappingHelpView struct {
	Width       int
	ColumnWidth int

	help.Model
}

func New() *WrappingHelpView {
	return &WrappingHelpView{
		Model: help.New(),
	}
}

func (h WrappingHelpView) View(keys KeyMap) string {
	if !h.ShowAll {
		return h.Model.View(keys)
	}
	return h.fullHelpViewWithWrap(keys.FullHelp())
}

// fullHelpViewWithWrap renders help columns with wrapping support.
// It maintains column structure but wraps columns to new rows when needed.
func (h WrappingHelpView) fullHelpViewWithWrap(groups [][]key.Binding) string {
	if len(groups) == 0 {
		return ""
	}

	separator := h.Styles.FullSeparator.Inline(true).Render(h.FullSeparator)
	sepWidth := lipgloss.Width(separator)

	// First pass: render all columns with fixed width and track heights
	var columns []string

	for _, group := range groups {
		if group == nil || !shouldRenderColumn(group) {
			continue
		}

		var keys []string
		var descriptions []string

		for _, kb := range group {
			if !kb.Enabled() {
				continue
			}
			keys = append(keys, kb.Help().Key)
			descriptions = append(descriptions, kb.Help().Desc)
		}

		// Build column with fixed width
		col := lipgloss.JoinHorizontal(lipgloss.Top,
			h.Styles.FullKey.Render(strings.Join(keys, "\n")),
			" ",
			h.Styles.FullDesc.Render(strings.Join(descriptions, "\n")),
		)

		col = lipgloss.NewStyle().Width(h.ColumnWidth).Render(col)

		columns = append(columns, col)
	}

	// Second pass: arrange columns into rows with wrapping
	maxCols := min((h.Width + sepWidth) / h.ColumnWidth)

	var rows []string
	for i := 0; i < len(columns); i += maxCols {
		cols := columns[i:min(i+maxCols, len(columns))]

		row := lipgloss.JoinHorizontal(lipgloss.Top, cols...)

		rows = append(rows, row)
	}

	return strings.Join(rows, "\n\n")
}

// shouldRenderColumn checks if a column has any enabled bindings
func shouldRenderColumn(b []key.Binding) bool {
	for _, v := range b {
		if v.Enabled() {
			return true
		}
	}
	return false
}
