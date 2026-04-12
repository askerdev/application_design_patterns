package tui

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var menuItems = []list.Item{
	menuItem{label: "Tasks",     dest: screenTasks},
	menuItem{label: "Projects",  dest: screenProjects},
	menuItem{label: "Notes",     dest: screenNotes},
	menuItem{label: "Reminders", dest: screenReminders},
	menuItem{label: "Tags",      dest: screenTags},
	menuItem{label: "Pomodoro",  dest: screenPomodoro},
	menuItem{label: "Stats",     dest: screenStats},
}

type menuItem struct {
	label string
	dest  screen
}

func (i menuItem) Title() string       { return i.label }
func (i menuItem) Description() string { return "" }
func (i menuItem) FilterValue() string { return i.label }

type homeModel struct {
	list list.Model
}

func newHomeModel() homeModel {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(colorPrimary).
		BorderLeftForeground(colorPrimary)

	l := list.New(menuItems, delegate, 0, 0)
	l.Title = "TaskFlow"
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)
	return homeModel{list: l}
}

func (m homeModel) Init() tea.Cmd { return nil }

func (m homeModel) Update(msg tea.Msg) (homeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "enter", "l":
			if item, ok := m.list.SelectedItem().(menuItem); ok {
				return m, func() tea.Msg { return navigateMsg{to: item.dest} }
			}
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m homeModel) View() string {
	return lipgloss.NewStyle().Padding(1, 2).Render(m.list.View())
}

func (m homeModel) setSize(w, h int) homeModel {
	m.list.SetSize(w-4, h-2)
	return m
}
