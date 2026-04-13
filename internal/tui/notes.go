package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	domain "taskflow/internal"
)

type noteItem struct {
	note        *domain.Note
	projectName string
	tagName     string
}

func (i noteItem) Title() string {
	return fmt.Sprintf("%s [%s]", i.note.Title, i.projectName)
}

func (i noteItem) Description() string {
	preview := i.note.Content
	if len(preview) > 60 {
		preview = preview[:60] + "…"
	}
	if i.tagName != "" {
		return preview + " #" + i.tagName
	}
	return preview
}

func (i noteItem) FilterValue() string {
	return i.note.Title + " " + i.projectName + " " + i.tagName
}

type notesModel struct {
	list     list.Model
	form     *huh.Form
	fTitle   *string
	fContent *string
	fProject *string
	fTag     *string

	svcs   Services
	user   *domain.User
	status string
	width  int
	height int
}

func newNotesModel(svcs Services, user *domain.User) notesModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Notes"
	l.Styles.Title = titleStyle
	return notesModel{svcs: svcs, user: user, list: l}
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
		notes, _ := m.svcs.Notes.List(m.user.ID)
		projects, _ := m.svcs.Projects.List(m.user.ID)
		projMap := make(map[int64]string, len(projects))
		for _, p := range projects {
			projMap[p.ID] = p.Name
		}
		tags, _ := m.svcs.Tags.List(m.user.ID)
		tagMap := make(map[int64]string, len(tags))
		for _, t := range tags {
			tagMap[t.ID] = t.Name
		}
		items := make([]list.Item, len(notes))
		for i, n := range notes {
			proj, tag := "", ""
			if n.ProjectID != nil {
				proj = projMap[*n.ProjectID]
			}
			if n.TagID != nil {
				tag = tagMap[*n.TagID]
			}
			items[i] = noteItem{note: n, projectName: proj, tagName: tag}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Unfiltered {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "a":
				projects, _ := m.svcs.Projects.List(m.user.ID)
				if len(projects) == 0 {
					m.status = errorStyle.Render("Create a project first.")
					return m, nil
				}
				title, content, proj, tag := "", "", fmt.Sprintf("%d", projects[0].ID), ""
				m.fTitle, m.fContent, m.fProject, m.fTag = &title, &content, &proj, &tag
				m.form = m.buildAddForm(projects)
				return m, m.form.Init()
			case "d":
				if item, ok := m.list.SelectedItem().(noteItem); ok {
					if err := m.svcs.Notes.Delete(item.note.ID); err != nil {
						m.status = errorStyle.Render("Error: " + err.Error())
					} else {
						m.status = "Deleted."
					}
					return m, m.reload()
				}
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
	help := statusBarStyle.Render("a add  d delete  / filter  tab switch  q quit")
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

func (m *notesModel) buildAddForm(projects []*domain.Project) *huh.Form {
	projOpts := make([]huh.Option[string], len(projects))
	for i, p := range projects {
		projOpts[i] = huh.NewOption(p.Name, fmt.Sprintf("%d", p.ID))
	}
	tags, _ := m.svcs.Tags.List(m.user.ID)
	tagOpts := []huh.Option[string]{huh.NewOption("(none)", "")}
	for _, t := range tags {
		tagOpts = append(tagOpts, huh.NewOption(t.Name, fmt.Sprintf("%d", t.ID)))
	}
	return huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Title").Value(m.fTitle),
			huh.NewText().Title("Content").Value(m.fContent),
			huh.NewSelect[string]().Title("Project").Options(projOpts...).Value(m.fProject),
			huh.NewSelect[string]().Title("Tag (optional)").Options(tagOpts...).Value(m.fTag),
		),
	)
}

func (m *notesModel) saveNote() error {
	var projID int64
	fmt.Sscanf(*m.fProject, "%d", &projID)
	var tagID *int64
	if *m.fTag != "" {
		id := int64(0)
		fmt.Sscanf(*m.fTag, "%d", &id)
		tagID = &id
	}
	factory := domain.NewEntityFactory()
	n := factory.CreateNote(m.user.ID, *m.fTitle, strings.TrimSpace(*m.fContent), &projID, tagID)
	return m.svcs.Notes.Create(n)
}
