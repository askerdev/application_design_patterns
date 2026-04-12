package task

import (
	"sync"

	"taskflow/internal/domain"
	"taskflow/internal/repository"
)

// CachingTaskRepo is the Proxy pattern: same interface as TaskRepository,
// adds transparent in-memory caching of per-user task lists.
type CachingTaskRepo struct {
	real  repository.TaskRepository
	mu    sync.Mutex
	cache map[int64][]*domain.Task // keyed by userID
}

func NewCachingTaskRepo(real repository.TaskRepository) *CachingTaskRepo {
	return &CachingTaskRepo{real: real, cache: make(map[int64][]*domain.Task)}
}

func (c *CachingTaskRepo) invalidate(userID int64) {
	c.mu.Lock()
	delete(c.cache, userID)
	c.mu.Unlock()
}

func (c *CachingTaskRepo) Create(t *domain.Task) error {
	err := c.real.Create(t)
	if err == nil {
		c.invalidate(t.UserID)
	}
	return err
}

func (c *CachingTaskRepo) GetByID(id int64) (*domain.Task, error) {
	return c.real.GetByID(id)
}

func (c *CachingTaskRepo) GetAllByUser(userID int64) ([]*domain.Task, error) {
	c.mu.Lock()
	if cached, ok := c.cache[userID]; ok {
		c.mu.Unlock()
		return cached, nil
	}
	c.mu.Unlock()

	tasks, err := c.real.GetAllByUser(userID)
	if err != nil {
		return nil, err
	}
	c.mu.Lock()
	c.cache[userID] = tasks
	c.mu.Unlock()
	return tasks, nil
}

func (c *CachingTaskRepo) GetByProject(projectID int64) ([]*domain.Task, error) {
	return c.real.GetByProject(projectID)
}

func (c *CachingTaskRepo) Update(t *domain.Task) error {
	err := c.real.Update(t)
	if err == nil {
		c.invalidate(t.UserID)
	}
	return err
}

func (c *CachingTaskRepo) Delete(id int64) error {
	// Need to look up userID to invalidate — fetch first.
	t, err := c.real.GetByID(id)
	if err == nil {
		defer c.invalidate(t.UserID)
	}
	return c.real.Delete(id)
}
