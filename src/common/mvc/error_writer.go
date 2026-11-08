package mvc

import (
	"net/http"
	"strings"

	"github.com/6022-labs/agentic-x402-facilitator/src/common/errors"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

// ErrorWriter writes and logs every error response; a 500 or 502 logs its cause and never writes it to the body.
type ErrorWriter struct {
	logger *zap.Logger
}

func NewErrorWriter(logger *zap.Logger) *ErrorWriter {
	return &ErrorWriter{logger: logger}
}

func (w *ErrorWriter) Write(ctx fiber.Ctx, err error) error {
	switch e := err.(type) {
	case *errors.ValidationError:
		return w.write(ctx, fiber.StatusBadRequest, "", e, nil)
	case *errors.BadRequestError:
		return w.write(ctx, fiber.StatusBadRequest, e.Code, e, nil)
	case *errors.UnavailableError:
		return w.write(ctx, fiber.StatusBadGateway, e.Code, e, e.Cause)
	case *errors.InternalError:
		return w.write(ctx, fiber.StatusInternalServerError, e.Code, e, e.Cause)
	case *fiber.Error:
		code := w.fiberErrorCode(e.Code)
		return w.write(ctx, e.Code, code, fiber.Map{"code": code, "message": e.Message}, nil)
	default:
		return w.Write(ctx, errors.NewInternalError("internal_error", err))
	}
}

func (w *ErrorWriter) write(ctx fiber.Ctx, status int, code string, body any, cause error) error {
	if err := ctx.Status(status).JSON(body); err != nil {
		return err
	}
	// 404s are routine probes and would drown the log.
	if status == fiber.StatusNotFound {
		return nil
	}
	fields := []zap.Field{
		zap.String("method", ctx.Method()),
		zap.String("path", ctx.OriginalURL()),
		zap.Int("status", status),
		zap.String("code", code),
		zap.ByteString("response", ctx.Response().Body()),
		zap.NamedError("cause", cause),
	}
	if status >= fiber.StatusInternalServerError {
		w.logger.Error("request failed", fields...)
	} else {
		w.logger.Warn("request rejected", fields...)
	}
	return nil
}

// fiberErrorCode turns a status into a snake_case code ("Request Entity Too Large" → request_entity_too_large).
func (w *ErrorWriter) fiberErrorCode(status int) string {
	text := http.StatusText(status)
	if text == "" {
		return "http_error"
	}
	return strings.ReplaceAll(strings.ToLower(text), " ", "_")
}
