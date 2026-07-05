package service

import (
	"context"

	"oat431/tiny-mchawa-api/internal/model"
)

// TodolistRepo defines the data access methods todolist service needs.
type TodolistRepo interface {
	Create(ctx context.Context, tl *model.Todolist) error
	GetByID(ctx context.Context, id string) (*model.Todolist, error)
	List(ctx context.Context, params model.ListTodolistsParams, ownedBy string) ([]model.Todolist, int, error)
	Update(ctx context.Context, id, title, description string) (*model.Todolist, error)
	Delete(ctx context.Context, id string) (bool, error)
	TaskStatusSummary(ctx context.Context, todolistID string) (total int, doneCount int, inprogressCount int, err error)
}

// TaskRepo defines the data access methods task service needs.
type TaskRepo interface {
	Create(ctx context.Context, t *model.Task) error
	GetByID(ctx context.Context, id string) (*model.Task, error)
	List(ctx context.Context, todolistID string, params model.ListTasksParams) ([]model.Task, int, error)
	Update(ctx context.Context, id, title, description, status string) (*model.Task, error)
	Delete(ctx context.Context, id string) (bool, error)
}
