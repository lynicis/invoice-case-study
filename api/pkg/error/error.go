package error

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type CustomError struct {
	error
	Code     int
	Message  string
	Severity zapcore.Level
	Fields   []zapcore.Field
}

const ContextKeyLog = "log"

func ErrorHandler(ctx *fiber.Ctx, err error) error {
	var cerr CustomError
	ok := errors.As(err, &cerr)
	if !ok {
		fiberError := err.(*fiber.Error)
		return ctx.SendStatus(fiberError.Code)
	}

	var log *zap.Logger
	log, ok = ctx.Locals(ContextKeyLog).(*zap.Logger)
	if !ok {
		return ctx.SendStatus(cerr.Code)
	}

	if len(cerr.Fields) > 0 {
		for _, field := range cerr.Fields {
			log = log.With(field)
		}
	}

	log.Log(cerr.Severity, cerr.Message)
	return ctx.SendStatus(cerr.Code)
}
