package task_test

import (
	"testing"

	"taskflow/internal/domain"
	"taskflow/internal/task"
)

// countingRepo wraps stubService to count GetAllByUser calls.
type countingRepo struct {
	calls int
	inner *stubService
}

func (r *countingRepo) Create(t *domain.Task) error { return r.inner.Create(t) }
func (r *countingRepo) GetByID(id int64) (*domain.Task, error) {
	if t, ok := r.inner.tasks[id]; ok {
		return t, nil
	}
	return nil, nil
}
func (r *countingRepo) GetAllByUser(userID int64) ([]*domain.Task, error) {
	r.calls++
	var result []*domain.Task
	for _, t := range r.inner.tasks {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}
func (r *countingRepo) GetByProject(projectID int64) ([]*domain.Task, error) { return nil, nil }
func (r *countingRepo) Update(t *domain.Task) error                          { return r.inner.Update(t) }
func (r *countingRepo) Delete(id int64) error                                { return r.inner.Delete(id) }

func TestCachingTaskRepo_CachesGetAllByUser(t *testing.T) {
	inner := newStubService()
	inner.tasks[1] = &domain.Task{ID: 1, UserID: 1, Content: "A"}
	repo := &countingRepo{inner: inner}
	proxy := task.NewCachingTaskRepo(repo)

	// First call hits real repo.
	if _, err := proxy.GetAllByUser(1); err != nil {
		t.Fatal(err)
	}
	// Second call should hit cache — real repo not called again.
	if _, err := proxy.GetAllByUser(1); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 1 {
		t.Errorf("expected 1 real call, got %d", repo.calls)
	}
}

func TestCachingTaskRepo_InvalidatesOnCreate(t *testing.T) {
	inner := newStubService()
	repo := &countingRepo{inner: inner}
	proxy := task.NewCachingTaskRepo(repo)

	proxy.GetAllByUser(1) // prime cache
	proxy.Create(&domain.Task{ID: 2, UserID: 1, Content: "B"})
	proxy.GetAllByUser(1) // should hit real repo again

	if repo.calls != 2 {
		t.Errorf("expected 2 real calls after Create invalidation, got %d", repo.calls)
	}
}
