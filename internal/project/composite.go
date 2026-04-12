package project

import "fmt"

// WorkItem is the component interface — implemented by both composite and leaf nodes.
type WorkItem interface {
	Title() string
	Description() string
	Children() []WorkItem
}

// ProjectNode is the composite — it has children (tasks, notes).
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

// TaskLeaf is a leaf node representing a single task.
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

// NoteLeaf is a leaf node representing a single note.
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

// Render formats the whole tree as a string for display.
// Composite items show their children indented.
func Render(item WorkItem, indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}
	result := prefix + item.Title() + "\n"
	for _, child := range item.Children() {
		result += Render(child, indent+1)
	}
	return result
}
