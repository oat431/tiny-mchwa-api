package model

import "time"

type Task struct {
	ID          string    `json:"id" db:"id"`
	TodolistID  string    `json:"todolistId" db:"todolist_id"`
	Title       string    `json:"title" db:"title"`
	Description *string   `json:"description" db:"description"`
	Status      string    `json:"status" db:"status"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type CreateTaskRequest struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
}

type UpdateTaskRequest struct {
	Title       string `json:"title" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Status      string `json:"status" validate:"required,oneof=pending inprogress done"`
}

type ListTasksParams struct {
	Page    int    `query:"page"`
	PerPage int    `query:"perPage"`
	Status  string `query:"status"`
	Title   string `query:"title"`
}

// ValidTaskStatus checks if a status is valid
func ValidTaskStatus(s string) bool {
	return s == "pending" || s == "inprogress" || s == "done"
}
