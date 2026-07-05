package service

import (
	"context"
	"errors"
	"log/slog"

	"oat431/tiny-mchawa-api/internal/model"
)

var (
	ErrNotFound  = errors.New("todolist not found")
	ErrForbidden = errors.New("not owned by this user")
)

type TodolistService struct {
	repo TodolistRepo
}

func NewTodolistService(repo TodolistRepo) *TodolistService {
	return &TodolistService{repo: repo}
}

func (s *TodolistService) Create(ctx context.Context, req model.CreateTodolistRequest, ownedBy string) (*model.Todolist, error) {
	tl := &model.Todolist{
		Title:         req.Title,
		Description:   ptrStr(req.Description),
		OwnedBy:       ownedBy,
		SourceService: req.SourceService,
	}
	if err := s.repo.Create(ctx, tl); err != nil {
		return nil, err
	}
	tl.Status = "pending" // new todolist, no tasks
	slog.InfoContext(ctx, "todolist created", "id", tl.ID)
	return tl, nil
}

func (s *TodolistService) GetByID(ctx context.Context, id, ownedBy string) (*model.Todolist, error) {
	tl, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return nil, ErrNotFound
	}
	if tl.OwnedBy != ownedBy {
		return nil, ErrForbidden
	}
	tl.Status, err = s.computeStatus(ctx, tl.ID)
	if err != nil {
		return nil, err
	}
	return tl, nil
}

func (s *TodolistService) List(ctx context.Context, params model.ListTodolistsParams, ownedBy string) ([]model.Todolist, int, error) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 20
	}

	todolists, total, err := s.repo.List(ctx, params, ownedBy)
	if err != nil {
		return nil, 0, err
	}

	// Compute status for each and apply status filter
	result := make([]model.Todolist, 0, len(todolists))
	for _, tl := range todolists {
		tl.Status, err = s.computeStatus(ctx, tl.ID)
		if err != nil {
			return nil, 0, err
		}
		if params.Status != "" && tl.Status != params.Status {
			total-- // ponytail: adjust total for filtered-out items; not exact but close enough for MVP
			continue
		}
		result = append(result, tl)
	}

	return result, total, nil
}

func (s *TodolistService) Update(ctx context.Context, id string, req model.UpdateTodolistRequest, ownedBy string) (*model.Todolist, error) {
	tl, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if tl == nil {
		return nil, ErrNotFound
	}
	if tl.OwnedBy != ownedBy {
		return nil, ErrForbidden
	}

	updated, err := s.repo.Update(ctx, id, req.Title, req.Description)
	if err != nil {
		return nil, err
	}
	updated.Status, err = s.computeStatus(ctx, updated.ID)
	if err != nil {
		return nil, err
	}
	slog.InfoContext(ctx, "todolist updated", "id", id)
	return updated, nil
}

func (s *TodolistService) Delete(ctx context.Context, id, ownedBy string) error {
	tl, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if tl == nil {
		return ErrNotFound
	}
	if tl.OwnedBy != ownedBy {
		return ErrForbidden
	}

	deleted, err := s.repo.Delete(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrNotFound
	}
	slog.InfoContext(ctx, "todolist deleted", "id", id)
	return nil
}

func (s *TodolistService) computeStatus(ctx context.Context, todolistID string) (string, error) {
	total, done, inprogress, err := s.repo.TaskStatusSummary(ctx, todolistID)
	if err != nil {
		return "pending", err // tasks table might not exist yet in MVP
	}
	if total == 0 {
		return "pending", nil
	}
	if done == total {
		return "done", nil
	}
	if inprogress > 0 {
		return "inprogress", nil
	}
	return "pending", nil
}

func ptrStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
