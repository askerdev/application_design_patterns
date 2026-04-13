package domain_test

import (
	"testing"

	"taskflow/internal"
)

// countingTaskRepo wraps stubTaskService to count GetAllByUser calls.
type countingTaskRepo struct {
	calls int
	inner *stubTaskService
}

func (r *countingTaskRepo) Create(t *domain.Task) error { return r.inner.Create(t) }
func (r *countingTaskRepo) GetByID(id int64) (*domain.Task, error) {
	if t, ok := r.inner.tasks[id]; ok {
		return t, nil
	}
	return nil, nil
}
func (r *countingTaskRepo) GetAllByUser(userID int64) ([]*domain.Task, error) {
	r.calls++
	var result []*domain.Task
	for _, t := range r.inner.tasks {
		if t.UserID == userID {
			result = append(result, t)
		}
	}
	return result, nil
}
func (r *countingTaskRepo) GetByProject(projectID int64) ([]*domain.Task, error) { return nil, nil }
func (r *countingTaskRepo) Update(t *domain.Task) error                          { return r.inner.Update(t) }
func (r *countingTaskRepo) Delete(id int64) error                                { return r.inner.Delete(id) }

func TestCachingTaskRepo_CachesGetAllByUser(t *testing.T) {
	inner := newStubTaskService()
	inner.tasks[1] = &domain.Task{ID: 1, UserID: 1, Content: "A"}
	repo := &countingTaskRepo{inner: inner}
	proxy := domain.NewCachingTaskRepo(repo)

	if _, err := proxy.GetAllByUser(1); err != nil {
		t.Fatal(err)
	}
	if _, err := proxy.GetAllByUser(1); err != nil {
		t.Fatal(err)
	}
	if repo.calls != 1 {
		t.Errorf("expected 1 real call, got %d", repo.calls)
	}
}

func TestCachingTaskRepo_InvalidatesOnCreate(t *testing.T) {
	inner := newStubTaskService()
	repo := &countingTaskRepo{inner: inner}
	proxy := domain.NewCachingTaskRepo(repo)

	proxy.GetAllByUser(1)
	proxy.Create(&domain.Task{ID: 2, UserID: 1, Content: "B"})
	proxy.GetAllByUser(1)

	if repo.calls != 2 {
		t.Errorf("expected 2 real calls after Create invalidation, got %d", repo.calls)
	}
}
