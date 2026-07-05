package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"oat431/tiny-mchawa-api/internal/model"

	"github.com/jmoiron/sqlx"
)

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, t *model.Task) error {
	query := `INSERT INTO tasks (todolist_id, title, description, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		t.TodolistID, t.Title, t.Description, t.Status,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*model.Task, error) {
	var t model.Task
	err := r.db.GetContext(ctx, &t,
		`SELECT id, todolist_id, title, description, status, created_at, updated_at
		 FROM tasks WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TaskRepository) List(ctx context.Context, todolistID string, params model.ListTasksParams) ([]model.Task, int, error) {
	where := []string{"todolist_id = $1"}
	args := []any{todolistID}
	argIdx := 2

	if params.Status != "" {
		where = append(where, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, params.Status)
		argIdx++
	}
	if params.Title != "" {
		where = append(where, fmt.Sprintf("title ILIKE $%d", argIdx))
		args = append(args, "%"+params.Title+"%")
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tasks WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (params.Page - 1) * params.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT id, todolist_id, title, description, status, created_at, updated_at
		 FROM tasks WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.TodolistID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *TaskRepository) Update(ctx context.Context, id, title, description, status string) (*model.Task, error) {
	var t model.Task
	err := r.db.QueryRowContext(ctx,
		`UPDATE tasks SET title = $1, description = $2, status = $3, updated_at = NOW()
		 WHERE id = $4
		 RETURNING id, todolist_id, title, description, status, created_at, updated_at`,
		title, description, status, id,
	).Scan(&t.ID, &t.TodolistID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

func (r *TaskRepository) Delete(ctx context.Context, id string) (bool, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}
