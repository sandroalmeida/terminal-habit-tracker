package ui

import "github.com/charmbracelet/lipgloss"

var (
	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1)

	ListHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true).
			Padding(0, 1)

	ListItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedListItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				PaddingLeft(2).
				Bold(true)

	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			MarginTop(1)
)
