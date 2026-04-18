package domain

import "sync"

type CachingTaskRepo struct {
	real  TaskRepository
	mu    sync.Mutex
	cache map[int64][]*Task
}

func NewCachingTaskRepo(real TaskRepository) *CachingTaskRepo {
	return &CachingTaskRepo{real: real, cache: make(map[int64][]*Task)}
}

func (c *CachingTaskRepo) invalidate(userID int64) {
	c.mu.Lock()
	delete(c.cache, userID)
	c.mu.Unlock()
}

func (c *CachingTaskRepo) Create(t *Task) error {
	err := c.real.Create(t)
	if err == nil {
		c.invalidate(t.UserID)
	}
	return err
}

func (c *CachingTaskRepo) GetByID(id int64) (*Task, error) {
	return c.real.GetByID(id)
}

func (c *CachingTaskRepo) GetAllByUser(userID int64) ([]*Task, error) {
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

func (c *CachingTaskRepo) GetByProject(projectID int64) ([]*Task, error) {
	return c.real.GetByProject(projectID)
}

func (c *CachingTaskRepo) Update(t *Task) error {
	err := c.real.Update(t)
	if err == nil {
		c.invalidate(t.UserID)
	}
	return err
}

func (c *CachingTaskRepo) Delete(id int64) error {
	t, err := c.real.GetByID(id)
	if err == nil {
		defer c.invalidate(t.UserID)
	}
	return c.real.Delete(id)
}
