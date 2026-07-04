package handler

import (
	"errors"

	"oat431/tiny-mchawa-api/internal/model"
	"oat431/tiny-mchawa-api/internal/service"

	"github.com/gofiber/fiber/v3"
)

type TodolistHandler struct {
	svc *service.TodolistService
}

func NewTodolistHandler(svc *service.TodolistService) *TodolistHandler {
	return &TodolistHandler{svc: svc}
}

// ponytail: hardcoded user for MVP — auth middleware will set this later
func getUserID(c fiber.Ctx) string {
	if uid, ok := c.Locals("userID").(string); ok && uid != "" {
		return uid
	}
	return "00000000-0000-0000-0000-000000000001"
}

func ok(c fiber.Ctx, data any, meta *model.PaginationMeta) error {
	return c.Status(fiber.StatusOK).JSON(model.APIResponse{Data: data, Meta: meta})
}

func created(c fiber.Ctx, data any) error {
	return c.Status(fiber.StatusCreated).JSON(model.APIResponse{Data: data})
}

func apiErr(c fiber.Ctx, status int, msg string, details ...string) error {
	return c.Status(status).JSON(model.APIResponse{
		Error: model.APIError{Code: status, Message: msg, Details: details},
	})
}

func (h *TodolistHandler) Create(c fiber.Ctx) error {
	var req model.CreateTodolistRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apiErr(c, 400, "Validation failed", err.Error())
	}
	if req.Title == "" {
		return apiErr(c, 400, "Validation failed", "title is required")
	}
	if req.SourceService == "" {
		return apiErr(c, 400, "Validation failed", "sourceService is required")
	}

	tl, err := h.svc.Create(c.Context(), req, getUserID(c))
	if err != nil {
		return apiErr(c, 500, "Internal server error")
	}
	return created(c, tl)
}

func (h *TodolistHandler) GetByID(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apiErr(c, 400, "ID is required")
	}

	tl, err := h.svc.GetByID(c.Context(), id, getUserID(c))
	if err != nil {
		return handleServiceErr(c, err)
	}
	return ok(c, tl, nil)
}

func (h *TodolistHandler) List(c fiber.Ctx) error {
	var params model.ListTodolistsParams
	if err := c.Bind().Query(&params); err != nil {
		return apiErr(c, 400, "Invalid query parameters")
	}

	todolists, total, err := h.svc.List(c.Context(), params, getUserID(c))
	if err != nil {
		return apiErr(c, 500, "Internal server error")
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
	return ok(c, todolists, meta)
}

func (h *TodolistHandler) Update(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apiErr(c, 400, "ID is required")
	}

	var req model.UpdateTodolistRequest
	if err := c.Bind().JSON(&req); err != nil {
		return apiErr(c, 400, "Validation failed", err.Error())
	}
	if req.Title == "" {
		return apiErr(c, 400, "Validation failed", "title is required")
	}

	tl, err := h.svc.Update(c.Context(), id, req, getUserID(c))
	if err != nil {
		return handleServiceErr(c, err)
	}
	return ok(c, tl, nil)
}

func (h *TodolistHandler) Delete(c fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return apiErr(c, 400, "ID is required")
	}

	if err := h.svc.Delete(c.Context(), id, getUserID(c)); err != nil {
		return handleServiceErr(c, err)
	}
	return c.SendStatus(204)
}

func handleServiceErr(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, service.ErrNotFound):
		return apiErr(c, 404, "Todolist not found")
	case errors.Is(err, service.ErrForbidden):
		return apiErr(c, 403, "Not owned by this user")
	default:
		return apiErr(c, 500, "Internal server error")
	}
}
