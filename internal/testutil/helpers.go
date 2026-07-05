package testutil

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"oat431/tiny-mchawa-api/internal/handler"
	mw "oat431/tiny-mchawa-api/internal/middleware"
	"oat431/tiny-mchawa-api/internal/model"
	"oat431/tiny-mchawa-api/internal/service"

	"github.com/gofiber/fiber/v3"
)

// --- Mock todolist store ---

type MockTodolistStore struct {
	mu     sync.RWMutex
	items  map[string]*model.Todolist
	nextID int
}

func NewMockTodolistStore() *MockTodolistStore {
	return &MockTodolistStore{items: make(map[string]*model.Todolist), nextID: 1}
}

func (s *MockTodolistStore) Create(tl *model.Todolist) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tl.ID = fmt.Sprintf("tl-%d", s.nextID)
	s.nextID++
	now := time.Now().Truncate(time.Second)
	tl.CreatedAt = now
	tl.UpdatedAt = now
	s.items[tl.ID] = tl
}

func (s *MockTodolistStore) Get(id string) *model.Todolist {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items[id]
}

func (s *MockTodolistStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

// --- Mock task store ---

type MockTaskStore struct {
	mu     sync.RWMutex
	items  map[string]*model.Task
	nextID int
}

func NewMockTaskStore() *MockTaskStore {
	return &MockTaskStore{items: make(map[string]*model.Task), nextID: 1}
}

func (s *MockTaskStore) Create(t *model.Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t.ID = fmt.Sprintf("task-%d", s.nextID)
	s.nextID++
	now := time.Now().Truncate(time.Second)
	t.CreatedAt = now
	t.UpdatedAt = now
	s.items[t.ID] = t
}

func (s *MockTaskStore) Get(id string) *model.Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.items[id]
}

func (s *MockTaskStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	return true
}

// --- Mock repositories (implement service interfaces) ---

type MockTodoRepo struct {
	Store     *MockTodolistStore
	TaskStore *MockTaskStore // for TaskStatusSummary
}

func (r *MockTodoRepo) Create(_ context.Context, tl *model.Todolist) error {
	r.Store.Create(tl)
	return nil
}

func (r *MockTodoRepo) GetByID(_ context.Context, id string) (*model.Todolist, error) {
	tl := r.Store.Get(id)
	return tl, nil // nil if not found
}

func (r *MockTodoRepo) List(_ context.Context, params model.ListTodolistsParams, ownedBy string) ([]model.Todolist, int, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()
	var result []model.Todolist
	for _, tl := range r.Store.items {
		if tl.OwnedBy != ownedBy {
			continue
		}
		if params.SourceService != "" && tl.SourceService != params.SourceService {
			continue
		}
		if params.Title != "" && !strings.Contains(strings.ToLower(tl.Title), strings.ToLower(params.Title)) {
			continue
		}
		result = append(result, *tl)
	}
	total := len(result)
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	start := (params.Page - 1) * params.PerPage
	if start >= total {
		return []model.Todolist{}, total, nil
	}
	end := start + params.PerPage
	if end > total {
		end = total
	}
	return result[start:end], total, nil
}

func (r *MockTodoRepo) Update(_ context.Context, id, title, description string) (*model.Todolist, error) {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()
	tl, ok := r.Store.items[id]
	if !ok {
		return nil, nil
	}
	tl.Title = title
	if description != "" {
		tl.Description = &description
	}
	tl.UpdatedAt = time.Now().Truncate(time.Second)
	return tl, nil
}

func (r *MockTodoRepo) Delete(_ context.Context, id string) (bool, error) {
	return r.Store.Delete(id), nil
}

func (r *MockTodoRepo) TaskStatusSummary(_ context.Context, todolistID string) (int, int, int, error) {
	if r.TaskStore == nil {
		return 0, 0, 0, nil
	}
	r.TaskStore.mu.RLock()
	defer r.TaskStore.mu.RUnlock()
	total, done, inprogress := 0, 0, 0
	for _, t := range r.TaskStore.items {
		if t.TodolistID == todolistID {
			total++
			if t.Status == "done" {
				done++
			}
			if t.Status == "inprogress" {
				inprogress++
			}
		}
	}
	return total, done, inprogress, nil
}

type MockTaskRepo struct {
	Store    *MockTaskStore
	TodoRepo *MockTodoRepo
}

func (r *MockTaskRepo) Create(_ context.Context, t *model.Task) error {
	r.Store.Create(t)
	return nil
}

func (r *MockTaskRepo) GetByID(_ context.Context, id string) (*model.Task, error) {
	t := r.Store.Get(id)
	return t, nil
}

func (r *MockTaskRepo) List(_ context.Context, todolistID string, params model.ListTasksParams) ([]model.Task, int, error) {
	r.Store.mu.RLock()
	defer r.Store.mu.RUnlock()
	var result []model.Task
	for _, t := range r.Store.items {
		if t.TodolistID != todolistID {
			continue
		}
		if params.Status != "" && t.Status != params.Status {
			continue
		}
		if params.Title != "" && !strings.Contains(strings.ToLower(t.Title), strings.ToLower(params.Title)) {
			continue
		}
		result = append(result, *t)
	}
	total := len(result)
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 20
	}
	start := (params.Page - 1) * params.PerPage
	if start >= total {
		return []model.Task{}, total, nil
	}
	end := start + params.PerPage
	if end > total {
		end = total
	}
	return result[start:end], total, nil
}

func (r *MockTaskRepo) Update(_ context.Context, id, title, description, status string) (*model.Task, error) {
	r.Store.mu.Lock()
	defer r.Store.mu.Unlock()
	t, ok := r.Store.items[id]
	if !ok {
		return nil, nil
	}
	t.Title = title
	if description != "" {
		t.Description = &description
	}
	t.Status = status
	t.UpdatedAt = time.Now().Truncate(time.Second)
	return t, nil
}

func (r *MockTaskRepo) Delete(_ context.Context, id string) (bool, error) {
	return r.Store.Delete(id), nil
}

// --- Test app builder ---

const TestUserID = "00000000-0000-0000-0000-000000000001"

// NewTestApp creates a Fiber app with mock services for handler testing.
func NewTestApp() (*fiber.App, *MockTodolistStore, *MockTaskStore) {
	todoStore := NewMockTodolistStore()
	taskStore := NewMockTaskStore()

	todoRepo := &MockTodoRepo{Store: todoStore, TaskStore: taskStore}
	taskRepo := &MockTaskRepo{Store: taskStore, TodoRepo: todoRepo}
	todoSvc := service.NewTodolistService(todoRepo)
	taskSvc := service.NewTaskService(taskRepo, todoRepo)

	todoH := handler.NewTodolistHandler(todoSvc)
	taskH := handler.NewTaskHandler(taskSvc)

	app := fiber.New(fiber.Config{
		StructValidator: mw.NewValidator(),
		ErrorHandler: func(c fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{"error": err.Error()})
		},
	})

	v1 := app.Group("/api/v1")
	v1.Post("/todolists", todoH.Create)
	v1.Get("/todolists", todoH.List)
	v1.Get("/todolists/:id", todoH.GetByID)
	v1.Put("/todolists/:id", todoH.Update)
	v1.Delete("/todolists/:id", todoH.Delete)
	v1.Post("/todolists/:todolistId/tasks", taskH.Create)
	v1.Get("/todolists/:todolistId/tasks", taskH.List)
	v1.Get("/tasks/:id", taskH.GetByID)
	v1.Put("/tasks/:id", taskH.Update)
	v1.Delete("/tasks/:id", taskH.Delete)

	app.Get("/health", func(c fiber.Ctx) error { return c.SendStatus(200) })
	app.Use(func(c fiber.Ctx) error {
		return c.Status(404).JSON(fiber.Map{"error": "not found"})
	})

	return app, todoStore, taskStore
}
