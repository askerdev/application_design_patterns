package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	domain "taskflow/internal/domain"
	"taskflow/internal/web"
)

type ganttProjectItem struct{ project *domain.Project }

func (i ganttProjectItem) Title() string       { return i.project.Name }
func (i ganttProjectItem) Description() string { return string(i.project.Status) }
func (i ganttProjectItem) FilterValue() string { return i.project.Name }

type ganttModel struct {
	list    list.Model
	spinner spinner.Model
	loading bool

	svcs   Services
	user   *domain.User
	server *web.GanttServer

	status string
	width  int
	height int

	// form-поле нужно для совместимости с activeHasForm в app.go (см. ниже).
	// Тут оно всегда nil.
}

type ganttPlanReadyMsg struct {
	plan *domain.GanttPlan
	url  string
	err  error
}

func newGanttModel(svcs Services, user *domain.User, server *web.GanttServer) ganttModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Gantt — выбор проекта"
	l.Styles.Title = titleStyle
	l.SetFilteringEnabled(false)

	sp := spinner.New()
	sp.Spinner = spinner.Dot

	return ganttModel{svcs: svcs, user: user, server: server, list: l, spinner: sp}
}

func (m ganttModel) reload() tea.Cmd {
	return func() tea.Msg { return ganttLoadMsg{} }
}

type ganttLoadMsg struct{}

func (m ganttModel) Init() tea.Cmd { return m.reload() }

func (m ganttModel) Update(msg tea.Msg) (ganttModel, tea.Cmd) {
	switch msg := msg.(type) {
	case ganttLoadMsg:
		projects, _ := m.svcs.Projects.List(m.user.ID)
		items := make([]list.Item, len(projects))
		for i, p := range projects {
			items[i] = ganttProjectItem{project: p}
		}
		m.list.SetItems(items)
		return m, nil

	case ganttPlanReadyMsg:
		m.loading = false
		if msg.err != nil {
			m.status = errorStyle.Render("Ошибка генерации: " + msg.err.Error())
			return m, nil
		}
		via := "fallback"
		if msg.plan.UsedLLM {
			via = "LLM"
		}
		m.status = fmt.Sprintf("✅ План готов (%s, задач: %d). Открыто: %s",
			via, len(msg.plan.Items), msg.url)
		return m, nil

	case spinner.TickMsg:
		if m.loading {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil

	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
		switch msg.String() {
		case "q":
			return m, tea.Quit
		case "g", "enter":
			return m.generate()
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m ganttModel) generate() (ganttModel, tea.Cmd) {
	item, ok := m.list.SelectedItem().(ganttProjectItem)
	if !ok {
		m.status = errorStyle.Render("Выберите проект.")
		return m, nil
	}
	if m.svcs.GanttPlanner == nil || m.server == nil {
		m.status = errorStyle.Render("Gantt-планировщик не инициализирован.")
		return m, nil
	}
	pid := item.project.ID
	m.loading = true
	m.status = "Генерация таймлайна... (модель LLM может думать до 3 минут)"

	planCmd := func() tea.Msg {
		plan, err := m.svcs.GanttPlanner.Plan(pid)
		if err != nil {
			return ganttPlanReadyMsg{err: err}
		}
		m.server.SetPlan(plan)
		if err := m.server.Start(); err != nil {
			return ganttPlanReadyMsg{plan: plan, err: fmt.Errorf("start server: %w", err)}
		}
		url := m.server.URL(pid)
		_ = web.OpenBrowser(url)
		return ganttPlanReadyMsg{plan: plan, url: url}
	}
	return m, tea.Batch(m.spinner.Tick, planCmd)
}

func (m ganttModel) View() string {
	body := strings.Builder{}
	body.WriteString(m.list.View())
	body.WriteString("\n")

	if m.loading {
		body.WriteString("\n" + m.spinner.View() + " " + m.status)
	} else if m.status != "" {
		body.WriteString("\n" + m.status)
	}

	if m.server != nil && m.server.IsRunning() {
		body.WriteString("\n" + lipgloss.NewStyle().Foreground(colorMuted).
			Render("HTTP сервер: http://"+m.server.Addr()))
	}

	help := statusBarStyle.Render("g/enter generate  tab switch  q quit")
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(body.String()),
		help,
	)
}

func (m ganttModel) setSize(w, h int) ganttModel {
	m.width, m.height = w, h
	m.list.SetSize(w-4, h-8)
	return m
}
