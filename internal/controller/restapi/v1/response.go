package v1

import "github.com/gofiber/fiber/v2"

// Response is the unified API response structure
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func errorResponse(c *fiber.Ctx, httpStatus int, message string) error {
	return c.Status(httpStatus).JSON(Response{
		Code: httpStatus,
		Msg:  message,
		Data: nil,
	})
}

func structuredErrorResponse(c *fiber.Ctx, httpStatus int, errCode string, message string) error {
	return c.Status(httpStatus).JSON(ErrorResponse{
		Error:   errCode,
		Message: message,
	})
}

func successResponse(c *fiber.Ctx, data interface{}) error {
	return c.JSON(Response{
		Code: fiber.StatusOK,
		Msg:  "success",
		Data: data,
	})
}

func successMessage(c *fiber.Ctx, message string) error {
	return c.JSON(Response{
		Code: fiber.StatusOK,
		Msg:  message,
		Data: nil,
	})
}

func createdResponse(c *fiber.Ctx, data interface{}) error {
	return c.Status(fiber.StatusCreated).JSON(Response{
		Code: fiber.StatusCreated,
		Msg:  "created",
		Data: data,
	})
}
