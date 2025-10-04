package sift

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	colorGreen       = lipgloss.Color("28")
	colorRed         = lipgloss.Color("124")
	colorMutedRed    = lipgloss.Color("#D25D5D")
	colorOrange      = lipgloss.Color("214")
	colorMutedOrange = lipgloss.Color("#D27E5D")
	colorBlue        = lipgloss.Color("27")
	colorMutedBlue   = lipgloss.Color("#2B57A3")
	colorGrey        = lipgloss.Color("244")

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

	styleLog = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("249"))

	styleHeader = lipgloss.NewStyle().Background(colorBlue).Bold(true).PaddingLeft(1).PaddingRight(1)

	styleBody = lipgloss.NewStyle().Padding(1)

	styleOutcome     = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1)
	styleOutcomePass = styleOutcome.Background(colorGreen)
	styleOutcomeFail = styleOutcome.Background(colorRed)
)
