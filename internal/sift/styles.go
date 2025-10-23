package sift

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
)

var (
	colorGreen = lipgloss.AdaptiveColor{
		Light: "#2D7F1E",
		Dark:  "#5FD700",
	}
	colorRed = lipgloss.AdaptiveColor{
		Light: "#C41E3A",
		Dark:  "#FF0000",
	}
	colorMutedRed = lipgloss.AdaptiveColor{
		Light: "#A04040",
		Dark:  "#D25D5D",
	}
	colorOrange = lipgloss.AdaptiveColor{
		Light: "#D97009",
		Dark:  "#FFAF00",
	}
	colorMutedOrange = lipgloss.AdaptiveColor{
		Light: "#A65D30",
		Dark:  "#D27E5D",
	}
	colorBlue = lipgloss.AdaptiveColor{
		Light: "#004080",
		Dark:  "#005FFF",
	}
	colorMutedBlue = lipgloss.AdaptiveColor{
		Light: "#4A90E2",
		Dark:  "#5B9BD5",
	}
	colorHighlight = lipgloss.AdaptiveColor{
		Light: "#E0E8F0",
		Dark:  "#2B57A3",
	}
	colorGrey = lipgloss.AdaptiveColor{
		Light: "#6C6C6C",
		Dark:  "#808080",
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
	styleHighlighted = lipgloss.NewStyle().Background(colorHighlight)

	styleLog = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{
		Light: "#4A4A4A",
		Dark:  "#B2B2B2",
	})

	styleHeader = lipgloss.NewStyle().Background(colorBlue).Bold(true).PaddingLeft(1).PaddingRight(1)

	styleBody = lipgloss.NewStyle().Padding(1)

	styleOutcome     = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	styleOutcomePass = styleOutcome.Background(colorGreen)
	styleOutcomeFail = styleOutcome.Background(colorRed)

	CenterDotPulse = spinner.Spinner{
		Frames: []string{
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#808080", Dark: "#4D4D4D"}).Render("∙"),
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A65D30", Dark: "#806040"}).Render("∙"),
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C97009", Dark: "#B38030"}).Render("∙"),
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#D97009", Dark: "#FFAF00"}).Render("∙"),
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#C97009", Dark: "#B38030"}).Render("∙"),
			lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#A65D30", Dark: "#806040"}).Render("∙"),
		},
		FPS: time.Second / 10,
	}
)
