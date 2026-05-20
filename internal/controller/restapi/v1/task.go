package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/usecase/task"
	"github.com/lminimum/LiteDock/pkg/logger"
)

type taskRoutes struct {
	t task.UseCase
	l logger.Interface
}

func NewTaskRoutes(handler fiber.Router, t *task.UseCase, l logger.Interface) {
	r := &taskRoutes{t: *t, l: l}

	h := handler.Group("/tasks")
	{
		h.Get("", r.list)
		h.Get("/:id", r.get)
	}
}

func (r *taskRoutes) list(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 20)
	offset := c.QueryInt("offset", 0)

	tasks, err := r.t.ListTasks(c.Context(), limit, offset)
	if err != nil {
		return errorResponse(c, fiber.StatusInternalServerError, "Failed to list tasks")
	}

	return successResponse(c, tasks)
}

func (r *taskRoutes) get(c *fiber.Ctx) error {
	id := c.Params("id")

	t, err := r.t.GetTask(c.Context(), id)
	if err != nil {
		return errorResponse(c, fiber.StatusNotFound, "Task not found")
	}

	return successResponse(c, t)
}
