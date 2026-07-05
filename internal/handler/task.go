package handler

import (
	"errors"

	"oat431/tiny-mchawa-api/internal/model"
	"oat431/tiny-mchawa-api/internal/service"

	"github.com/gofiber/fiber/v3"
)

type TaskHandler struct {
	svc *service.TaskService
}

func NewTaskHandler(svc *service.TaskService) *TaskHandler {
	return &TaskHandler{svc: svc}
}

func (h *TaskHandler) Create(c fiber.Ctx) error {
	todolistID := c.Params("todolistId")
	if todolistID == "" {
		return apiErr(c, 400, "todolistId is required")
	}

	var req model.CreateTaskRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apiErr(c, 400, "Validation failed", err.Error())
	}

	task, err := h.svc.Create(c.Context(), todolistID, req, getUserID(c))
	if err != nil {
		return handleTaskServiceErr(c, err)
	}
	return created(c, task)
}

func (h *TaskHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apiErr(c, 400, "ID is required")
	}

	task, err := h.svc.GetByID(c.Context(), id, getUserID(c))
	if err != nil {
		return handleTaskServiceErr(c, err)
	}
	return ok(c, task, nil)
}

func (h *TaskHandler) List(c fiber.Ctx) error {
	todolistID := c.Params("todolistId")
	if todolistID == "" {
		return apiErr(c, 400, "todolistId is required")
	}

	var params model.ListTasksParams
	if err := c.Bind().Query(&params); err != nil {
		return apiErr(c, 400, "Invalid query parameters")
	}

	tasks, total, err := h.svc.List(c.Context(), todolistID, params, getUserID(c))
	if err != nil {
		return handleTaskServiceErr(c, err)
	}

	perPage := params.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}
	page := params.Page
	if page < 1 {
		page = 1
	}

	meta := &model.PaginationMeta{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: (total + perPage - 1) / perPage,
	}
	return ok(c, tasks, meta)
}

func (h *TaskHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apiErr(c, 400, "ID is required")
	}

	var req model.UpdateTaskRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apiErr(c, 400, "Validation failed", err.Error())
	}

	task, err := h.svc.Update(c.Context(), id, req, getUserID(c))
	if err != nil {
		return handleTaskServiceErr(c, err)
	}
	return ok(c, task, nil)
}

func (h *TaskHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apiErr(c, 400, "ID is required")
	}

	if err := h.svc.Delete(c.Context(), id, getUserID(c)); err != nil {
		return handleTaskServiceErr(c, err)
	}
	return c.SendStatus(204)
}

func handleTaskServiceErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return apiErr(c, 404, "Task not found")
	case errors.Is(err, service.ErrForbidden):
		return apiErr(c, 403, "Not owned by this user")
	default:
		return apiErr(c, 500, "Internal server error")
	}
}
