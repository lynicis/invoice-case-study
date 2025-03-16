package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"

	"invoice-api/internal/invoice"
	"invoice-api/pkg/config"
	customError "invoice-api/pkg/error"
)

type GlobalHandler interface {
	RegisterRoutes()
}

func main() {
	cfg := config.Read()
	log, err := zap.NewProduction()
	if err != nil {
		panic("failed to initialize logger")
	}
	defer log.Sync()

	invoicePgRepository := invoice.NewPgRepository(
		log,
		cfg.Postgresql.Host,
		cfg.Postgresql.Port,
		cfg.Postgresql.Username,
		cfg.Postgresql.Password,
		cfg.Postgresql.Database,
	)

	server := fiber.New(fiber.Config{
		JSONDecoder:           json.Unmarshal,
		JSONEncoder:           json.Marshal,
		ErrorHandler:          customError.ErrorHandler,
		DisableStartupMessage: true,
	})
	server.Use(func(ctx *fiber.Ctx) error {
		ctx.Locals(customError.ContextKeyLog, log)
		return ctx.Next()
	})
	server.Use(recover.New())
	server.Use(cors.New(cors.Config{AllowOrigins: cfg.CorsOrigins}))
	server.Use(pprof.New())
	server.Get("/metrics", monitor.New())

	validate := validator.New()
	handlers := []GlobalHandler{invoice.NewHandler(server, validate, invoicePgRepository)}
	for _, handler := range handlers {
		handler.RegisterRoutes()
	}

	go func() {
		if err = server.Listen(fmt.Sprintf("0.0.0.0:%s", cfg.ServerPort)); err != nil {
			log.Fatal("failed to start server", zap.Error(err))
		}
	}()

	log.Info("server started on port", zap.String("port", cfg.ServerPort))
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	<-signalChannel

	log.Info("shutting down server...")
	if err = server.ShutdownWithTimeout(5 * time.Second); err != nil {
		log.Fatal("error occurred while server shutdown", zap.Error(err))
	}
}
