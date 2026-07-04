package model

import "time"

type Todolist struct {
	ID            string    `json:"id" db:"id"`
	Title         string    `json:"title" db:"title"`
	Description   *string   `json:"description" db:"description"`
	Status        string    `json:"status" db:"-"` // computed, not stored
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
	OwnedBy       string    `json:"ownedBy" db:"owned_by"`
	SourceService string    `json:"sourceService" db:"source_service"`
}

type CreateTodolistRequest struct {
	Title         string `json:"title" validate:"required,max=255"`
	Description   string `json:"description" validate:"max=1000"`
	SourceService string `json:"sourceService" validate:"required,max=100"`
}

type UpdateTodolistRequest struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type ListTodolistsParams struct {
	Page          int    `query:"page"`
	PerPage       int    `query:"perPage"`
	Status        string `query:"status"`
	SourceService string `query:"sourceService"`
	Title         string `query:"title"`
}

// Pagination meta
type PaginationMeta struct {
	Page       int `json:"page"`
	PerPage    int `json:"perPage"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// Standard API response
type APIResponse struct {
	Data  any             `json:"data"`
	Error any             `json:"error"`
	Meta  *PaginationMeta `json:"meta"`
}
