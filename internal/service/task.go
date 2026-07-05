package service

import (
	"context"
	"log/slog"

	"oat431/tiny-mchawa-api/internal/model"
	"oat431/tiny-mchawa-api/internal/repository"
)

type TaskService struct {
	taskRepo    *repository.TaskRepository
	todoRepo    *repository.TodolistRepository
	todoService *TodolistService
}

func NewTaskService(taskRepo *repository.TaskRepository, todoRepo *repository.TodolistRepository, todoService *TodolistService) *TaskService {
	return &TaskService{taskRepo: taskRepo, todoRepo: todoRepo, todoService: todoService}
}

func (s *TaskService) Create(ctx context.Context, todolistID string, req model.CreateTaskRequest, ownedBy string) (*model.Task, error) {
	// Verify todolist exists and belongs to user
	tl, err := s.todoRepo.GetByID(ctx, todolistID)
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return nil, ErrNotFound
	}
	if tl.OwnedBy != ownedBy {
		return nil, ErrForbidden
	}

	task := &model.Task{
		TodolistID:  todolistID,
		Title:       req.Title,
		Description: ptrStr(req.Description),
		Status:      "pending",
	}
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "task created", "id", task.ID, "todolistId", todolistID)
	return task, nil
}

func (s *TaskService) GetByID(ctx context.Context, id, ownedBy string) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}
	// Verify ownership via parent todolist
	tl, err := s.todoRepo.GetByID(ctx, task.TodolistID)
	if err != nil {
		return nil, err
	}
	if tl == nil || tl.OwnedBy != ownedBy {
		return nil, ErrForbidden
	}
	return task, nil
}

func (s *TaskService) List(ctx context.Context, todolistID string, params model.ListTasksParams, ownedBy string) ([]model.Task, int, error) {
	// Verify todolist ownership
	tl, err := s.todoRepo.GetByID(ctx, todolistID)
	if err != nil {
		return nil, 0, err
	}
	if tl == nil {
		return nil, 0, ErrNotFound
	}
	if tl.OwnedBy != ownedBy {
		return nil, 0, ErrForbidden
	}

	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	return s.taskRepo.List(ctx, todolistID, params)
}

func (s *TaskService) Update(ctx context.Context, id string, req model.UpdateTaskRequest, ownedBy string) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, ErrNotFound
	}
	// Verify ownership via parent todolist
	tl, err := s.todoRepo.GetByID(ctx, task.TodolistID)
	if err != nil {
		return nil, err
	}
	if tl == nil || tl.OwnedBy != ownedBy {
		return nil, ErrForbidden
	}

	updated, err := s.taskRepo.Update(ctx, id, req.Title, req.Description, req.Status)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "task updated", "id", id)
	return updated, nil
}

func (s *TaskService) Delete(ctx context.Context, id, ownedBy string) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if task == nil {
		return ErrNotFound
	}
	// Verify ownership via parent todolist
	tl, err := s.todoRepo.GetByID(ctx, task.TodolistID)
	if err != nil {
		return err
	}
	if tl == nil || tl.OwnedBy != ownedBy {
		return ErrForbidden
	}

	deleted, err := s.taskRepo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	slog.InfoContext(ctx, "task deleted", "id", id)
	return nil
}
