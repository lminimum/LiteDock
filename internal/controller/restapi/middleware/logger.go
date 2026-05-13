package middleware

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/lminimum/LiteDock/pkg/logger"
)

// sanitizeURL removes token query param from URL to prevent JWT leakage in logs.
func sanitizeURL(rawURL string) string {
	re := regexp.MustCompile(`(\?|&)token=[^&]*`)
	return re.ReplaceAllString(rawURL, "")
}

func buildRequestMessage(ctx *fiber.Ctx) string {
	var result strings.Builder

	result.WriteString(ctx.IP())
	result.WriteString(" - ")
	result.WriteString(ctx.Method())
	result.WriteString(" ")
	result.WriteString(sanitizeURL(ctx.OriginalURL()))
	result.WriteString(" - ")
	result.WriteString(strconv.Itoa(ctx.Response().StatusCode()))
	result.WriteString(" ")
	result.WriteString(strconv.Itoa(len(ctx.Response().Body())))

	return result.String()
}

func Logger(l logger.Interface) func(c *fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		err := ctx.Next()

		l.Debug(buildRequestMessage(ctx))

		return err
	}
}
