package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.Color("69")  // soft purple
	colorSecondary = lipgloss.Color("212") // pink
	colorMuted     = lipgloss.Color("241")
	colorError     = lipgloss.Color("196")
	colorSuccess   = lipgloss.Color("78")
	colorOverdue   = lipgloss.Color("196")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("230")).
			Background(lipgloss.Color("238")).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().Foreground(colorError)

	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary)

	inactivePaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted)

	overdueStyle = lipgloss.NewStyle().Foreground(colorOverdue).Bold(true)
)
