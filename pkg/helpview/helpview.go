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
	var columnHeights []int

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

		// Pad column to fixed width
		lines := strings.Split(col, "\n")
		var paddedLines []string
		for _, line := range lines {
			paddedLine := lipgloss.NewStyle().Width(h.ColumnWidth).Render(line)
			paddedLines = append(paddedLines, paddedLine)
		}
		
		columns = append(columns, strings.Join(paddedLines, "\n"))
		columnHeights = append(columnHeights, len(paddedLines))
	}

	// Second pass: arrange columns into rows with wrapping
	var rows []string
	var currentRow []string
	var currentRowHeights []int
	currentRowWidth := 0

	for i, col := range columns {
		neededWidth := h.ColumnWidth
		if len(currentRow) > 0 {
			neededWidth += sepWidth
		}

		// Check if adding this column would exceed width
		if len(currentRow) > 0 && currentRowWidth+neededWidth > h.Width {
			// Render current row and start new one
			rows = append(rows, h.renderRow(currentRow, currentRowHeights, separator))
			currentRow = []string{}
			currentRowHeights = []int{}
			currentRowWidth = 0
		}

		// Add column to current row
		currentRow = append(currentRow, col)
		currentRowHeights = append(currentRowHeights, columnHeights[i])
		currentRowWidth += h.ColumnWidth
		if len(currentRow) > 1 {
			currentRowWidth += sepWidth
		}
	}

	// Add final row
	if len(currentRow) > 0 {
		rows = append(rows, h.renderRow(currentRow, currentRowHeights, separator))
	}

	return strings.Join(rows, "\n\n")
}

// renderRow renders a row of columns, padding shorter columns to match the tallest
func (h WrappingHelpView) renderRow(columns []string, heights []int, separator string) string {
	if len(columns) == 0 {
		return ""
	}
	if len(columns) == 1 {
		return columns[0]
	}

	// Find max height
	maxHeight := 0
	for _, height := range heights {
		if height > maxHeight {
			maxHeight = height
		}
	}

	// Pad all columns to same height
	paddedColumns := make([]string, len(columns))
	for i, col := range columns {
		lines := strings.Split(col, "\n")
		// Pad with empty lines if needed
		for len(lines) < maxHeight {
			lines = append(lines, lipgloss.NewStyle().Width(h.ColumnWidth).Render(""))
		}
		paddedColumns[i] = strings.Join(lines, "\n")
	}

	// Join with separator
	result := []string{}
	for i, col := range paddedColumns {
		if i > 0 {
			result = append(result, separator)
		}
		result = append(result, col)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, result...)
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
