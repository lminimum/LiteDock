package v1

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/internal/action"
	"github.com/lminimum/LiteDock/internal/mcp"
)

type MCPHandler struct {
	mcpHandler *mcp.Handler
}

func NewMCPRoutes(apiV1Group fiber.Router, mcpHandler *mcp.Handler) {
	h := &MCPHandler{
		mcpHandler: mcpHandler,
	}

	apiV1Group.Post("/mcp", h.HandleMCP)
}

func (h *MCPHandler) HandleMCP(c *fiber.Ctx) error {
	var req mcp.MCPRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(mcp.MCPResponse{
			JSONRPC: "2.0",
			Error: &mcp.MCPError{
				Code:    -32700,
				Message: "Parse error: " + err.Error(),
			},
		})
	}

	userID, _ := c.Locals("userID").(string)
	sessionID, _ := c.Locals("sessionID").(string)

	ctx := c.UserContext()
	if userID != "" {
		ctx = context.WithValue(ctx, action.CtxKeyUserID, userID)
	}
	if sessionID != "" {
		ctx = context.WithValue(ctx, action.CtxKeySessionID, sessionID)
	}

	resp := h.mcpHandler.Handle(ctx, &req)
	return c.JSON(resp)
}
