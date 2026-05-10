package tui

import "github.com/charmbracelet/lipgloss"

var (
	colorPrimary   = lipgloss.AdaptiveColor{Light: "25", Dark: "69"} // тёмно-синий в светлой
	colorSecondary = lipgloss.AdaptiveColor{Light: "204", Dark: "212"} // чуть темнее розовый в светлой
	colorMuted     = lipgloss.AdaptiveColor{Light: "246", Dark: "241"} // серый чуть темнее в светлой
	colorError     = lipgloss.AdaptiveColor{Light: "160", Dark: "196"} // тёмно-красный в светлой
	colorSuccess   = lipgloss.AdaptiveColor{Light: "34", Dark: "78"}  // тёмно-зеленый в светлой
	colorOverdue   = lipgloss.AdaptiveColor{Light: "160", Dark: "196"}

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "230"}).
			Background(lipgloss.AdaptiveColor{Light: "240", Dark: "238"}).
			Padding(0, 1)

	errorStyle = lipgloss.NewStyle().Foreground(colorError)

	activePaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary)

	inactivePaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorMuted)

	overdueStyle = lipgloss.NewStyle().Foreground(colorOverdue).Bold(true)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.AdaptiveColor{Light: "15", Dark: "0"}). // белый текст на темном фоне
			Background(colorPrimary).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(colorMuted).
				Padding(0, 1)
)
