package repository

import domain "taskflow/internal"

// Type aliases re-export domain's repository interfaces.
// The canonical definitions live in taskflow/internal/domain.
type TaskRepository = domain.TaskRepository
type ProjectRepository = domain.ProjectRepository
type NoteRepository = domain.NoteRepository
type ReminderRepository = domain.ReminderRepository
type PomodoroRepository = domain.PomodoroRepository
type TagRepository = domain.TagRepository
type UserRepository = domain.UserRepository
