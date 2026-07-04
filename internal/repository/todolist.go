package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"oat431/tiny-mchawa-api/internal/model"

	"github.com/jmoiron/sqlx"
)

type TodolistRepository struct {
	db *sqlx.DB
}

func NewTodolistRepository(db *sqlx.DB) *TodolistRepository {
	return &TodolistRepository{db: db}
}

func (r *TodolistRepository) Create(ctx context.Context, tl *model.Todolist) error {
	query := `INSERT INTO todolists (title, description, owned_by, source_service)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		tl.Title, tl.Description, tl.OwnedBy, tl.SourceService,
	).Scan(&tl.ID, &tl.CreatedAt, &tl.UpdatedAt)
}

func (r *TodolistRepository) GetByID(ctx context.Context, id string) (*model.Todolist, error) {
	var tl model.Todolist
	err := r.db.GetContext(ctx, &tl,
		`SELECT id, title, description, created_at, updated_at, owned_by, source_service
		 FROM todolists WHERE id = $1`, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tl, err
}

func (r *TodolistRepository) List(ctx context.Context, params model.ListTodolistsParams, ownedBy string) ([]model.Todolist, int, error) {
	where := []string{"owned_by = $1"}
	args := []any{ownedBy}
	argIdx := 2

	if params.SourceService != "" {
		where = append(where, fmt.Sprintf("source_service = $%d", argIdx))
		args = append(args, params.SourceService)
		argIdx++
	}
	if params.Title != "" {
		where = append(where, fmt.Sprintf("title ILIKE $%d", argIdx))
		args = append(args, "%"+params.Title+"%")
		argIdx++
	}

	whereClause := strings.Join(where, " AND ")

	// Count total
	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM todolists WHERE %s", whereClause)
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// ponytail: status filter done in app layer after computing — SQL can't filter computed fields
	// Fetch page
	offset := (params.Page - 1) * params.PerPage
	dataQuery := fmt.Sprintf(
		`SELECT id, title, description, created_at, updated_at, owned_by, source_service
		 FROM todolists WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		whereClause, argIdx, argIdx+1)
	args = append(args, params.PerPage, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var todolists []model.Todolist
	for rows.Next() {
		var tl model.Todolist
		if err := rows.Scan(&tl.ID, &tl.Title, &tl.Description, &tl.CreatedAt, &tl.UpdatedAt, &tl.OwnedBy, &tl.SourceService); err != nil {
			return nil, 0, err
		}
		todolists = append(todolists, tl)
	}
	return todolists, total, rows.Err()
}

func (r *TodolistRepository) Update(ctx context.Context, id string, title, description string) (*model.Todolist, error) {
	var tl model.Todolist
	err := r.db.QueryRowContext(ctx,
		`UPDATE todolists SET title = $1, description = $2, updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, title, description, created_at, updated_at, owned_by, source_service`,
		title, description, id,
	).Scan(&tl.ID, &tl.Title, &tl.Description, &tl.CreatedAt, &tl.UpdatedAt, &tl.OwnedBy, &tl.SourceService)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tl, err
}

func (r *TodolistRepository) Delete(ctx context.Context, id string) (bool, error) {
	result, err := r.db.ExecContext(ctx, "DELETE FROM todolists WHERE id = $1", id)
	if err != nil {
		return false, err
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// CountTasks returns task count and status summary for a todolist.
// Used for computed status.
func (r *TodolistRepository) TaskStatusSummary(ctx context.Context, todolistID string) (total int, doneCount int, inprogressCount int, err error) {
	err = r.db.QueryRowContext(ctx,
		`SELECT COUNT(*), 
			COUNT(*) FILTER (WHERE status = 'done'),
			COUNT(*) FILTER (WHERE status = 'inprogress')
		 FROM tasks WHERE todolist_id = $1`, todolistID,
	).Scan(&total, &doneCount, &inprogressCount)
	return
}
