package domain

type Command interface {
	Execute() error
	Undo() error
}

type CommandHistory struct {
	stack []Command
}

func (h *CommandHistory) Push(c Command) {
	h.stack = append(h.stack, c)
}

func (h *CommandHistory) Undo() error {
	if len(h.stack) == 0 {
		return nil
	}
	last := h.stack[len(h.stack)-1]
	h.stack = h.stack[:len(h.stack)-1]
	return last.Undo()
}

type CompleteTaskCommand struct {
	svc        TaskService
	task       *Task
	prevStatus TaskStatus
}

func NewCompleteTaskCommand(svc TaskService, t *Task) *CompleteTaskCommand {
	return &CompleteTaskCommand{svc: svc, task: t, prevStatus: t.Status}
}

func (c *CompleteTaskCommand) Execute() error {
	c.task.Complete()
	return c.svc.Update(c.task)
}

func (c *CompleteTaskCommand) Undo() error {
	c.task.Status = c.prevStatus
	return c.svc.Update(c.task)
}

type DeleteTaskCommand struct {
	svc  TaskService
	task *Task
}

func NewDeleteTaskCommand(svc TaskService, t *Task) *DeleteTaskCommand {
	return &DeleteTaskCommand{svc: svc, task: t}
}

func (c *DeleteTaskCommand) Execute() error {
	return c.svc.Delete(c.task.ID)
}

func (c *DeleteTaskCommand) Undo() error {
	c.task.ID = 0
	return c.svc.Create(c.task)
}
