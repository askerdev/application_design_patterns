package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"taskflow/internal/domain"
)

type noteItem struct{ note *domain.Note }

func (i noteItem) Title() string {
	if i.note.ProjectID != nil {
		return fmt.Sprintf("%s [proj:%d]", i.note.Title, *i.note.ProjectID)
	}
	return i.note.Title
}

func (i noteItem) Description() string {
	if len(i.note.Content) > 60 {
		return i.note.Content[:60] + "…"
	}
	return i.note.Content
}

func (i noteItem) FilterValue() string { return i.note.Title }

type notesModel struct {
	list     list.Model
	form     *huh.Form
	fTitle   string
	fContent string
	fProject string
	fTag     string

	repos  *repos
	user   *domain.User
	status string
	width  int
	height int
}

func newNotesModel(r *repos, user *domain.User) notesModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Notes"
	l.Styles.Title = titleStyle
	return notesModel{repos: r, user: user, list: l}
}

func (m notesModel) reload() tea.Cmd {
	return func() tea.Msg { return notesLoadMsg{} }
}

type notesLoadMsg struct{}

func (m notesModel) Init() tea.Cmd { return m.reload() }

func (m notesModel) Update(msg tea.Msg) (notesModel, tea.Cmd) {
	if m.form != nil {
		f, cmd := m.form.Update(msg)
		if form, ok := f.(*huh.Form); ok {
			m.form = form
		}
		if m.form.State == huh.StateCompleted {
			m.form = nil
			if err := m.saveNote(); err != nil {
				m.status = errorStyle.Render("Error: " + err.Error())
			} else {
				m.status = "Note created."
			}
			return m, m.reload()
		}
		if m.form.State == huh.StateAborted {
			m.form = nil
			m.status = "Cancelled."
			return m, nil
		}
		return m, cmd
	}

	switch msg := msg.(type) {
	case notesLoadMsg:
		notes, _ := m.repos.notes.GetAllByUser(m.user.ID)
		items := make([]list.Item, len(notes))
		for i, n := range notes {
			items[i] = noteItem{note: n}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "h", "backspace":
			return m, func() tea.Msg { return backMsg{} }
		case "a":
			m.fTitle, m.fContent, m.fProject, m.fTag = "", "", "", ""
			m.form = m.buildAddForm()
			return m, m.form.Init()
		case "d":
			if item, ok := m.list.SelectedItem().(noteItem); ok {
				if err := m.repos.notes.Delete(item.note.ID); err != nil {
					m.status = errorStyle.Render("Error: " + err.Error())
				} else {
					m.status = "Deleted."
				}
				return m, m.reload()
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m notesModel) View() string {
	if m.form != nil {
		return lipgloss.NewStyle().Padding(1, 2).Render(m.form.View())
	}
	help := statusBarStyle.Render("a add  d delete  esc back")
	body := m.list.View()
	if m.status != "" {
		body += "\n" + m.status
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		lipgloss.NewStyle().Padding(1, 2).Render(body),
		help,
	)
}

func (m notesModel) setSize(w, h int) notesModel {
	m.width, m.height = w, h
	m.list.SetSize(w-4, h-4)
	return m
}

func (m *notesModel) buildAddForm() *huh.Form {
	projects, _ := m.repos.projects.GetAllByUser(m.user.ID)
	projOpts := []huh.Option[string]{huh.NewOption("(none)", "")}
	for _, p := range projects {
		projOpts = append(projOpts, huh.NewOption(p.Name, fmt.Sprintf("%d", p.ID)))
	}
	tags, _ := m.repos.tags.GetAllByUser(m.user.ID)
	tagOpts := []huh.Option[string]{huh.NewOption("(none)", "")}
	for _, t := range tags {
		tagOpts = append(tagOpts, huh.NewOption(t.Name, fmt.Sprintf("%d", t.ID)))
	}
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Title").Value(&m.fTitle),
			huh.NewText().Title("Content").Value(&m.fContent),
			huh.NewSelect[string]().Title("Project (optional)").Options(projOpts...).Value(&m.fProject),
			huh.NewSelect[string]().Title("Tag (optional)").Options(tagOpts...).Value(&m.fTag),
		),
	)
}

func (m *notesModel) saveNote() error {
	n := &domain.Note{
		UserID:  m.user.ID,
		Title:   m.fTitle,
		Content: strings.TrimSpace(m.fContent),
	}
	if m.fProject != "" {
		var id int64
		fmt.Sscanf(m.fProject, "%d", &id)
		n.ProjectID = &id
	}
	if m.fTag != "" {
		var id int64
		fmt.Sscanf(m.fTag, "%d", &id)
		n.TagID = &id
	}
	return m.repos.notes.Create(n)
}
