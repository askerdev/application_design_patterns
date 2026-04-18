package domain

import (
	"fmt"
	"strings"
)

type WorkItem interface {
	Title() string
	Description() string
	Children() []WorkItem
}

type ProjectNode struct {
	name     string
	status   string
	children []WorkItem
}

func NewProjectNode(name, status string) *ProjectNode {
	return &ProjectNode{name: name, status: status}
}

func (n *ProjectNode) Add(child WorkItem) {
	n.children = append(n.children, child)
}

func (n *ProjectNode) Title() string        { return fmt.Sprintf("[%s] %s", n.status, n.name) }
func (n *ProjectNode) Description() string  { return fmt.Sprintf("%d items", len(n.children)) }
func (n *ProjectNode) Children() []WorkItem { return n.children }

type TaskLeaf struct {
	content string
	status  string
	overdue bool
}

func NewTaskLeaf(content, status string, overdue bool) *TaskLeaf {
	return &TaskLeaf{content: content, status: status, overdue: overdue}
}

func (l *TaskLeaf) Title() string {
	s := fmt.Sprintf("[%s] %s", l.status, l.content)
	if l.overdue {
		s += " ⚠ OVERDUE"
	}
	return s
}
func (l *TaskLeaf) Description() string  { return "" }
func (l *TaskLeaf) Children() []WorkItem { return nil }

type NoteLeaf struct {
	title   string
	preview string
}

func NewNoteLeaf(title, content string) *NoteLeaf {
	preview := content
	if len(preview) > 40 {
		preview = preview[:40] + "…"
	}
	return &NoteLeaf{title: title, preview: preview}
}

func (l *NoteLeaf) Title() string        { return "📝 " + l.title }
func (l *NoteLeaf) Description() string  { return l.preview }
func (l *NoteLeaf) Children() []WorkItem { return nil }

func Render(item WorkItem, indent int) string {
	var prefix strings.Builder
	for range indent {
		prefix.WriteString("  ")
	}
	var result strings.Builder
	result.WriteString(prefix.String() + item.Title() + "\n")
	for _, child := range item.Children() {
		result.WriteString(Render(child, indent+1))
	}
	return result.String()
}
