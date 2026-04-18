package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	domain "taskflow/internal/domain"
)

type tagItem struct{ tag *domain.Tag }

func (i tagItem) Title() string       { return fmt.Sprintf("%s  %s", i.tag.Name, i.tag.Color) }
func (i tagItem) Description() string { return "" }
func (i tagItem) FilterValue() string { return i.tag.Name }

type tagsModel struct {
	list   list.Model
	form   *huh.Form
	fName  *string
	fColor *string

	svcs   Services
	user   *domain.User
	status string
	width  int
	height int
}

func newTagsModel(svcs Services, user *domain.User) tagsModel {
	l := list.New(nil, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Tags"
	l.Styles.Title = titleStyle
	return tagsModel{svcs: svcs, user: user, list: l}
}

func (m tagsModel) reload() tea.Cmd {
	return func() tea.Msg { return tagsLoadMsg{} }
}

type tagsLoadMsg struct{}

func (m tagsModel) Init() tea.Cmd { return m.reload() }

func (m tagsModel) Update(msg tea.Msg) (tagsModel, tea.Cmd) {
	if m.form != nil {
		if key, ok := msg.(tea.KeyMsg); ok && key.String() == "esc" {
			m.form = nil
			m.status = "Cancelled."
			return m, nil
		}
		f, cmd := m.form.Update(msg)
		if form, ok := f.(*huh.Form); ok {
			m.form = form
		}
		if m.form.State == huh.StateCompleted {
			m.form = nil
			if err := m.saveTag(); err != nil {
				m.status = errorStyle.Render("Error: " + err.Error())
			} else {
				m.status = "Tag created."
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
	case tagsLoadMsg:
		tags, _ := m.svcs.Tags.List(m.user.ID)
		items := make([]list.Item, len(tags))
		for i, t := range tags {
			items[i] = tagItem{tag: t}
		}
		m.list.SetItems(items)
		return m, nil

	case tea.KeyMsg:
		if m.list.FilterState() == list.Unfiltered {
			switch msg.String() {
			case "q":
				return m, tea.Quit
			case "a":
				name, color := "", "#ffffff"
				m.fName, m.fColor = &name, &color
				m.form = huh.NewForm(
					huh.NewGroup(
						huh.NewInput().Title("Name").Value(m.fName),
						huh.NewInput().Title("Color hex").Placeholder("#ffffff").Value(m.fColor),
					),
				)
				return m, m.form.Init()
			case "d":
				if item, ok := m.list.SelectedItem().(tagItem); ok {
					if err := m.svcs.Tags.Delete(item.tag.ID); err != nil {
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

func (m tagsModel) View() string {
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

func (m tagsModel) setSize(w, h int) tagsModel {
	m.width, m.height = w, h
	m.list.SetSize(w-4, h-4)
	return m
}

func (m *tagsModel) saveTag() error {
	factory := domain.NewEntityFactory()
	t := factory.CreateTag(m.user.ID, *m.fName, *m.fColor)
	return m.svcs.Tags.Create(t)
}
