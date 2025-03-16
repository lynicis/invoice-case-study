package invoice

import (
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go.uber.org/zap"

	customError "invoice-api/pkg/error"
)

type Handler struct {
	server     *fiber.App
	validator  *validator.Validate
	repository Repository
}

func NewHandler(server *fiber.App, validator *validator.Validate, repository Repository) *Handler {
	return &Handler{
		server:     server,
		validator:  validator,
		repository: repository,
	}
}

func (h *Handler) RegisterRoutes() {
	h.server.Post("/invoices", h.CreateInvoice)
	h.server.Get("/invoices", h.GetInvoices)
	h.server.Get("/invoices/:id", h.GetInvoiceById)
	h.server.Put("/invoices/:id", h.UpdateInvoiceById)
	h.server.Delete("/invoices/:id", h.DeleteInvoiceById)
}

func (h *Handler) CreateInvoice(ctx *fiber.Ctx) error {
	log := ctx.Locals(customError.ContextKeyLog).(*zap.Logger)
	log.With(zap.String("method", "CreateInvoice"))
	ctx.Locals(customError.ContextKeyLog, log)

	var reqBody CreateInvoiceRequest
	if err := ctx.BodyParser(&reqBody); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid request body",
			Severity: zap.WarnLevel,
		}
	}

	if err := h.validator.StructCtx(ctx.UserContext(), &reqBody); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid request body",
			Severity: zap.WarnLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	if err := h.repository.CreateInvoice(ctx.UserContext(), &InvoiceDTO{
		Id:          uuid.NewString(),
		ServiceName: reqBody.ServiceName,
		Amount:      reqBody.Amount,
		Status:      reqBody.Status,
		Date:        reqBody.Date,
	}); err != nil {
		return err
	}

	ctx.Locals(customError.ContextKeyLog).(*zap.Logger).Info("successfully finished")
	return ctx.SendStatus(fiber.StatusCreated)
}

func (h *Handler) GetInvoices(ctx *fiber.Ctx) error {
	log := ctx.Locals(customError.ContextKeyLog).(*zap.Logger)
	log.With(zap.String("method", "GetInvoices"))
	ctx.Locals(customError.ContextKeyLog, log)

	var queries GetInvoicesRequest
	if err := ctx.QueryParser(&queries); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid request query",
			Severity: zap.WarnLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	if err := h.validator.StructCtx(ctx.UserContext(), &queries); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid request query",
			Severity: zap.WarnLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	if queries.Page == 0 {
		queries.Page = 1
	}

	if queries.PageSize == 0 {
		queries.PageSize = 50
	}

	invoices, err := h.repository.GetInvoices(
		ctx.UserContext(),
		queries.Page,
		queries.PageSize,
		queries.Search,
	)
	if err != nil {
		return err
	}

	ctx.Locals(customError.ContextKeyLog).(*zap.Logger).Info("successfully finished")
	return ctx.JSON(invoices)
}

func (h *Handler) GetInvoiceById(ctx *fiber.Ctx) error {
	log := ctx.Locals(customError.ContextKeyLog).(*zap.Logger)
	log.With(zap.String("method", "GetInvoiceById"))
	ctx.Locals(customError.ContextKeyLog, log)

	invoiceId := ctx.Params("id")
	if err := h.validator.VarCtx(ctx.UserContext(), invoiceId, "required,uuid4"); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid invoice id",
			Severity: zap.WarnLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	invoice, err := h.repository.GetInvoiceById(ctx.UserContext(), invoiceId)
	if err != nil {
		return err
	}

	ctx.Locals(customError.ContextKeyLog).(*zap.Logger).Info("successfully finished")
	return ctx.JSON(invoice)
}

func (h *Handler) UpdateInvoiceById(ctx *fiber.Ctx) error {
	log := ctx.Locals(customError.ContextKeyLog).(*zap.Logger)
	log.With(zap.String("method", "UpdateInvoiceById"))
	ctx.Locals(customError.ContextKeyLog, log)

	var reqBody CreateInvoiceRequest
	if err := ctx.BodyParser(&reqBody); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid request body",
			Severity: zap.WarnLevel,
		}
	}

	invoiceId := ctx.Params("id")
	if err := h.validator.StructCtx(ctx.UserContext(), &UpdateInvoiceRequest{
		CreateInvoiceRequest: reqBody,
		Id:                   invoiceId,
	}); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid request body",
			Severity: zap.WarnLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	if err := h.repository.UpdateInvoiceById(ctx.UserContext(), invoiceId, &InvoiceDTO{
		Id:          invoiceId,
		ServiceName: reqBody.ServiceName,
		Amount:      reqBody.Amount,
		Status:      reqBody.Status,
	}); err != nil {
		return err
	}

	ctx.Locals(customError.ContextKeyLog).(*zap.Logger).Info("successfully finished")
	return ctx.SendStatus(fiber.StatusNoContent)
}

func (h *Handler) DeleteInvoiceById(ctx *fiber.Ctx) error {
	log := ctx.Locals(customError.ContextKeyLog).(*zap.Logger)
	log.With(zap.String("method", "DeleteInvoiceById"))
	ctx.Locals(customError.ContextKeyLog, log)

	id := ctx.Params("id")
	if err := h.validator.VarCtx(ctx.UserContext(), id, "required,uuid4"); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusBadRequest,
			Message:  "invalid invoice id",
			Severity: zap.WarnLevel,
		}
	}

	if err := h.repository.DeleteInvoiceById(ctx.UserContext(), id); err != nil {
		return err
	}

	ctx.Locals(customError.ContextKeyLog).(*zap.Logger).Info("successfully finished")
	return ctx.SendStatus(fiber.StatusNoContent)
}
