package tui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"taskflow/internal/domain"
	"taskflow/internal/pomodoro"
)

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg { return tickMsg(t) })
}

type sessionItem struct {
	s           *domain.PomodoroSession
	projectName string
}

func (i sessionItem) Title() string {
	start := ""
	if i.s.StartTime != nil {
		start = " @ " + i.s.StartTime.Format("2006-01-02 15:04")
	}
	return fmt.Sprintf("[%s] %dmin%s", i.s.State, i.s.WorkDuration, start)
}
func (i sessionItem) Description() string { return i.projectName }
func (i sessionItem) FilterValue() string { return string(i.s.State) }

type pomodoroModel struct {
	list      list.Model
	machine   *pomodoro.PomodoroMachine
	session   *domain.PomodoroSession
	progress  progress.Model
	totalSecs int
	form      *huh.Form
	fDur      *string
	fProject  *string

	svcs   Services
	user   *domain.User
	status string
	width  int
	height int
}

func newPomodoroModel(svcs Services, user *domain.User) pomodoroModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Pomodoro"
	l.Styles.Title = titleStyle
	return pomodoroModel{
		svcs:     svcs,
		user:     user,
		list:     l,
		progress: progress.New(progress.WithDefaultGradient()),
	}
}

func (m pomodoroModel) reload() tea.Cmd {
	return func() tea.Msg { return pomodoroLoadMsg{} }
}

type pomodoroLoadMsg struct{}

func (m pomodoroModel) Init() tea.Cmd { return m.reload() }

func (m pomodoroModel) Update(msg tea.Msg) (pomodoroModel, tea.Cmd) {
	if m.form != nil {
		f, cmd := m.form.Update(msg)
		if form, ok := f.(*huh.Form); ok {
			m.form = form
		}
		if m.form.State == huh.StateCompleted {
			m.form = nil
			return m, m.startSession()
		}
		if m.form.State == huh.StateAborted {
			m.form = nil
			m.status = "Cancelled."
			return m, nil
		}
		return m, cmd
	}

	if m.machine != nil {
		switch msg := msg.(type) {
		case tickMsg:
			if m.machine.StateName() == "RUNNING" {
				m.machine.RemainingTime--
				m.session.RemainingTime = m.machine.RemainingTime
				if m.machine.RemainingTime <= 0 {
					m.machine.Complete()
					m.syncSession()
					m.machine = nil
					m.status = "Session complete!"
					return m, m.reload()
				}
			}
			pct := 1.0 - float64(m.machine.RemainingTime)/float64(m.totalSecs)
			var progCmd tea.Cmd
			{
				updated, cmd := m.progress.Update(m.progress.SetPercent(pct))
				progCmd = cmd
				if p, ok := updated.(progress.Model); ok {
					m.progress = p
				}
			}
			return m, tea.Batch(progCmd, tickCmd())

		case tea.KeyMsg:
			switch msg.String() {
			case "p":
				m.machine.Pause()
				m.syncSession()
			case "r":
				m.machine.Resume()
				m.syncSession()
				return m, tickCmd()
			case "c":
				m.machine.Complete()
				m.syncSession()
				m.machine = nil
				m.status = "Session completed!"
				return m, m.reload()
			case "esc", "x":
				m.machine.Cancel()
				m.syncSession()
				m.machine = nil
				m.status = "Session cancelled."
				return m, m.reload()
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case pomodoroLoadMsg:
		sessions, _ := m.svcs.Pomodoro.List(m.user.ID)
		projects, _ := m.svcs.Projects.List(m.user.ID)
		projMap := make(map[int64]string, len(projects))
		for _, p := range projects {
			projMap[p.ID] = p.Name
		}
		items := make([]list.Item, len(sessions))
		for i, s := range sessions {
			name := ""
			if s.ProjectID != nil {
				name = projMap[*s.ProjectID]
			}
			items[i] = sessionItem{s: s, projectName: name}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			if m.machine == nil || m.machine.StateName() != "RUNNING" {
				return m, tea.Quit
			}
		case "a":
			projects, _ := m.svcs.Projects.List(m.user.ID)
			if len(projects) == 0 {
				m.status = errorStyle.Render("Create a project first.")
				return m, nil
			}
			dur, proj := "25", fmt.Sprintf("%d", projects[0].ID)
			m.fDur, m.fProject = &dur, &proj
			projOpts := make([]huh.Option[string], len(projects))
			for i, p := range projects {
				projOpts[i] = huh.NewOption(p.Name, fmt.Sprintf("%d", p.ID))
			}
			m.form = huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Work duration (minutes)").Placeholder("25").Value(m.fDur),
					huh.NewSelect[string]().Title("Project").Options(projOpts...).Value(m.fProject),
				),
			)
			return m, m.form.Init()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *pomodoroModel) startSession() tea.Cmd {
	dur := 25
	if m.fDur != nil {
		fmt.Sscanf(*m.fDur, "%d", &dur)
	}
	if dur <= 0 {
		dur = 25
	}
	m.totalSecs = dur * 60

	session := &domain.PomodoroSession{
		UserID:        m.user.ID,
		WorkDuration:  dur,
		RemainingTime: dur * 60,
		State:         domain.SessionStateIdle,
	}
	if m.fProject != nil {
		var id int64
		fmt.Sscanf(*m.fProject, "%d", &id)
		session.ProjectID = &id
	}
	if err := m.svcs.Pomodoro.Create(session); err != nil {
		m.status = errorStyle.Render("Error: " + err.Error())
		return nil
	}
	machine := pomodoro.NewPomodoroMachine(dur)
	machine.Start()
	session.State = domain.SessionState(machine.StateName())
	if err := m.svcs.Pomodoro.Update(session); err != nil {
		m.status = errorStyle.Render("Error: " + err.Error())
	}

	m.machine = machine
	m.session = session
	return tickCmd()
}

func (m *pomodoroModel) syncSession() {
	if m.session == nil || m.machine == nil {
		return
	}
	m.session.State = domain.SessionState(m.machine.StateName())
	m.session.RemainingTime = m.machine.RemainingTime
	if m.machine.StartTime != nil {
		m.session.StartTime = m.machine.StartTime
	}
	if m.machine.FinishTime != nil {
		m.session.FinishTime = m.machine.FinishTime
	}
	if err := m.svcs.Pomodoro.Update(m.session); err != nil {
		m.status = errorStyle.Render("Sync error: " + err.Error())
	}
}

func (m pomodoroModel) View() string {
	if m.form != nil {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.form.View())
	}

	if m.machine != nil {
		rem := m.machine.RemainingTime
		mins := rem / 60
		secs := rem % 60
		timerStr := lipgloss.NewStyle().Bold(true).Render(
			fmt.Sprintf("  %s   %02d:%02d\n\n", m.machine.StateName(), mins, secs),
		)
		help := statusBarStyle.Render("p pause  r resume  c complete  esc/x cancel")
		return lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Padding(2, 4).Render(
				timerStr+m.progress.View(),
			),
			help,
		)
	}

	help := statusBarStyle.Render("a new session  tab switch  q quit")
	body := m.list.View()
	if m.status != "" {
		body += "\n" + m.status
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(body),
		help,
	)
}

func (m pomodoroModel) setSize(w, h int) pomodoroModel {
	m.width, m.height = w, h
	m.list.SetSize(w-4, h-4)
	m.progress.Width = w - 12
	return m
}
