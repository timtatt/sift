package sift

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	colorGreen = lipgloss.AdaptiveColor{
		Light: "#2D7F1E", // Darker green for light mode
		Dark:  "#5FD700", // ANSI 28 equivalent bright green for dark mode
	}
	colorRed = lipgloss.AdaptiveColor{
		Light: "#C41E3A", // Darker red for light mode
		Dark:  "#FF0000", // ANSI 124 equivalent bright red for dark mode
	}
	colorMutedRed = lipgloss.AdaptiveColor{
		Light: "#A04040", // Darker muted red for light mode
		Dark:  "#D25D5D", // Original color for dark mode
	}
	colorOrange = lipgloss.AdaptiveColor{
		Light: "#D97009", // Darker orange for light mode
		Dark:  "#FFAF00", // ANSI 214 equivalent bright orange for dark mode
	}
	colorMutedOrange = lipgloss.AdaptiveColor{
		Light: "#A65D30", // Darker muted orange for light mode
		Dark:  "#D27E5D", // Original color for dark mode
	}
	colorBlue = lipgloss.AdaptiveColor{
		Light: "#004080", // Darker blue for light mode (header background)
		Dark:  "#005FFF", // ANSI 27 equivalent bright blue for dark mode
	}
	colorMutedBlue = lipgloss.AdaptiveColor{
		Light: "#E0E8F0", // Light blue/grey for light mode (highlight background)
		Dark:  "#2B57A3", // Original color for dark mode
	}
	colorGrey = lipgloss.AdaptiveColor{
		Light: "#6C6C6C", // Medium grey for light mode
		Dark:  "#808080", // ANSI 244 equivalent grey for dark mode
	}

	styleIcon = lipgloss.NewStyle().Bold(true)

	styleTick = styleIcon.
			Foreground(colorGreen)

	styleCross = styleIcon.
			Foreground(colorRed)

	styleProgress = styleIcon.
			Foreground(colorOrange)

	styleSkip = styleIcon.
			Foreground(colorMutedBlue)

	styleSecondary   = lipgloss.NewStyle().Foreground(colorGrey)
	styleHighlighted = lipgloss.NewStyle().Background(colorMutedBlue)

	styleLog = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.AdaptiveColor{
		Light: "#4A4A4A", // Darker grey for light mode
		Dark:  "#B2B2B2", // ANSI 249 equivalent for dark mode
	})

	styleHeader = lipgloss.NewStyle().Background(colorBlue).Bold(true).PaddingLeft(1).PaddingRight(1)

	styleBody = lipgloss.NewStyle().Padding(1)

	styleOutcome     = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	styleOutcomePass = styleOutcome.Background(colorGreen)
	styleOutcomeFail = styleOutcome.Background(colorRed)
)
