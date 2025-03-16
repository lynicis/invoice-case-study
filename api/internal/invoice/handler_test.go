package invoice

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	json "github.com/bytedance/sonic"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap"

	customError "invoice-api/pkg/error"
)

func TestHandler_NewHandler(t *testing.T) {
	h := NewHandler(nil, nil, nil)
	assert.NotNil(t, h)
}

func TestHandler_RegisterRoutes(t *testing.T) {
	h := NewHandler(fiber.New(), nil, nil)

	assert.NotPanics(t, h.RegisterRoutes)
}

func TestHandler_CreateInvoice(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run("happy path", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().CreateInvoice(gomock.Any(), gomock.Any()).Return(nil).Times(3)

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		requestBody := []CreateInvoiceRequest{
			{
				ServiceName: "DMP",
				Date:        time.Now().UTC(),
				Amount:      1,
				Status:      "PENDING",
			},
			{
				ServiceName: "SSP",
				Date:        time.Now().UTC(),
				Amount:      1,
				Status:      "PAID",
			},
			{
				ServiceName: "SSP",
				Date:        time.Now().UTC(),
				Amount:      1,
				Status:      "UNPAID",
			},
		}

		for _, body := range requestBody {
			marshalledReqBody, err := json.Marshal(body)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/invoices", strings.NewReader(string(marshalledReqBody)))
			assert.NoError(t, err)

			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

			res, err := server.Test(req, -1)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusCreated, res.StatusCode)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		server, validate := SetupServer(t)
		h := NewHandler(server, validate, nil)
		h.RegisterRoutes()

		requestBody := []interface{}{
			CreateInvoiceRequest{
				ServiceName: "INVALID",
				Amount:      1,
				Status:      "PAID",
				Date:        time.Now().UTC(),
			},
			CreateInvoiceRequest{
				ServiceName: "SSP",
				Amount:      0,
				Status:      "PAID",
				Date:        time.Now().UTC(),
			},
			CreateInvoiceRequest{
				ServiceName: "DMP",
				Amount:      1,
				Status:      "INVALID",
				Date:        time.Now().UTC(),
			},
			CreateInvoiceRequest{
				ServiceName: "DMP",
				Amount:      1,
				Status:      "PAID",
			},
		}
		for _, body := range requestBody {
			marshalledReqBody, err := json.Marshal(body)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/invoices", strings.NewReader(string(marshalledReqBody)))
			assert.NoError(t, err)

			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

			res, err := server.Test(req, -1)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().CreateInvoice(gomock.Any(), gomock.Any()).Return(customError.CustomError{
			Code:     http.StatusInternalServerError,
			Message:  "repository error",
			Severity: zap.ErrorLevel,
		})

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		reqBody := CreateInvoiceRequest{
			ServiceName: "DMP",
			Date:        time.Now().UTC(),
			Amount:      1,
			Status:      "PENDING",
		}

		marshalledReqBody, err := json.Marshal(reqBody)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPost, "/invoices", strings.NewReader(string(marshalledReqBody)))
		assert.NoError(t, err)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

		res, err := server.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})
}

func TestHandler_GetInvoices(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	invoices := &[]InvoiceDTO{
		{
			Id:          uuid.NewString(),
			ServiceName: "DMP",
			Amount:      1,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		},
		{
			Id:          uuid.NewString(),
			ServiceName: "DMP",
			Amount:      1,
			Status:      "PENDING",
			Date:        time.Now().UTC(),
		},
	}

	t.Run("happy path", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.
			EXPECT().
			GetInvoices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(invoices, nil).
			Times(6)

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		queries := []map[string]string{
			nil,
			{
				"page": "1",
			},
			{
				"pageSize": "1",
			},
			{
				"search": "test",
			},
			{
				"date": "ASC",
			},
			{
				"amount": "DESC",
			},
		}

		for _, query := range queries {
			reqUrl, err := url.Parse("http://0.0.0.0/invoices")
			assert.NoError(t, err)

			queryParams := reqUrl.Query()

			for queryKey, queryValue := range query {
				queryParams.Add(queryKey, queryValue)
			}

			reqUrl.RawQuery = queryParams.Encode()

			req := httptest.NewRequest(http.MethodGet, reqUrl.String(), nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

			res, err := server.Test(req, -1)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)
		}
	})

	t.Run("invalid request queries", func(t *testing.T) {
		server, validate := SetupServer(t)
		h := NewHandler(server, validate, nil)
		h.RegisterRoutes()

		queries := []map[string]string{
			{
				"date": "invalid",
			},
			{
				"amount": "invalid",
			},
		}

		for _, query := range queries {
			reqUrl, err := url.Parse("http://0.0.0.0/invoices")
			assert.NoError(t, err)

			queryParams := reqUrl.Query()

			for queryKey, queryValue := range query {
				queryParams.Add(queryKey, queryValue)
			}

			reqUrl.RawQuery = queryParams.Encode()

			req := httptest.NewRequest(http.MethodGet, reqUrl.String(), nil)
			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

			res, err := server.Test(req, -1)

			assert.NoError(t, err)
			assert.Equal(t, http.StatusBadRequest, res.StatusCode)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.
			EXPECT().
			GetInvoices(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(&[]InvoiceDTO{}, customError.CustomError{
				Code:     fiber.StatusInternalServerError,
				Message:  "repository error",
				Severity: zap.ErrorLevel,
			})

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		req := httptest.NewRequest(http.MethodGet, "/invoices", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

		res, err := server.Test(req, -1)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	})
}

func TestHandler_GetInvoiceById(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run("happy path", func(t *testing.T) {
		id := uuid.NewString()

		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().GetInvoiceById(gomock.Any(), gomock.Any()).Return(&InvoiceDTO{
			Id:          id,
			ServiceName: "DMP",
			Amount:      1,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		}, nil)

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/invoices/%s", id), nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		res, err := server.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusOK, res.StatusCode)
		assert.Equal(t, fiber.MIMEApplicationJSON, res.Header.Get(fiber.HeaderContentType))
	})

	t.Run("invalid invoice id", func(t *testing.T) {
		server, validate := SetupServer(t)
		h := NewHandler(server, validate, nil)
		h.RegisterRoutes()

		req := httptest.NewRequest(http.MethodGet, "/invoices/123", nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		res, err := server.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().GetInvoiceById(gomock.Any(), gomock.Any()).Return(nil, customError.CustomError{
			Code:     http.StatusInternalServerError,
			Message:  "repository error",
			Severity: zap.ErrorLevel,
		})

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/invoices/%s", uuid.NewString()), nil)
		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

		res, err := server.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, res.StatusCode)
	})
}

func TestHandler_UpdateInvoiceById(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run("happy path", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().UpdateInvoiceById(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(3)

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		requestBody := []CreateInvoiceRequest{
			{
				ServiceName: "DMP",
				Date:        time.Now().UTC(),
				Amount:      1,
				Status:      "PENDING",
			},
			{
				ServiceName: "SSP",
				Date:        time.Now().UTC(),
				Amount:      1,
				Status:      "PAID",
			},
			{
				ServiceName: "SSP",
				Date:        time.Now().UTC(),
				Amount:      1,
				Status:      "UNPAID",
			},
		}

		for _, body := range requestBody {
			marshalledReqBody, err := json.Marshal(body)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/invoices/%s", uuid.NewString()), strings.NewReader(string(marshalledReqBody)))
			assert.NoError(t, err)

			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

			res, err := server.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusNoContent, res.StatusCode)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		server, validate := SetupServer(t)
		h := NewHandler(server, validate, nil)
		h.RegisterRoutes()

		requestBody := []CreateInvoiceRequest{
			{},
			{
				ServiceName: "INVALID",
				Date:        time.Now().UTC(),
				Amount:      1,
				Status:      "PENDING",
			},
			{
				ServiceName: "DMP",
				Amount:      1,
				Status:      "PAID",
			},
			{
				ServiceName: "SSP",
				Date:        time.Now().UTC(),
				Amount:      0,
				Status:      "UNPAID",
			},
			{
				ServiceName: "SSP",
				Amount:      1,
				Date:        time.Now().UTC(),
				Status:      "INVALID",
			},
		}

		for _, body := range requestBody {
			marshalledReqBody, err := json.Marshal(body)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/invoices/%s", uuid.NewString()), strings.NewReader(string(marshalledReqBody)))
			assert.NoError(t, err)

			req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

			res, err := server.Test(req, -1)
			assert.NoError(t, err)
			assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().UpdateInvoiceById(gomock.Any(), gomock.Any(), gomock.Any()).Return(customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "repository error",
			Severity: zap.ErrorLevel,
		})

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		requestBody := CreateInvoiceRequest{
			ServiceName: "DMP",
			Date:        time.Now().UTC(),
			Amount:      1,
			Status:      "PENDING",
		}

		marshalledReqBody, err := json.Marshal(requestBody)
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("/invoices/%s", uuid.NewString()), strings.NewReader(string(marshalledReqBody)))
		assert.NoError(t, err)

		req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
		req.Header.Set(fiber.HeaderAccept, fiber.MIMEApplicationJSON)

		res, err := server.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, res.StatusCode)
	})
}

func TestHandler_DeleteInvoiceById(t *testing.T) {
	mockController := gomock.NewController(t)
	defer mockController.Finish()

	t.Run("happy path", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().DeleteInvoiceById(gomock.Any(), gomock.Any()).Return(nil)

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/invoices/%s", uuid.NewString()), nil)
		assert.NoError(t, err)

		res, err := server.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusNoContent, res.StatusCode)
	})

	t.Run("invalid request body", func(t *testing.T) {
		server, validate := SetupServer(t)
		h := NewHandler(server, validate, nil)
		h.RegisterRoutes()

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/invoices/%s", "invalid-id"), nil)
		assert.NoError(t, err)

		res, err := server.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusBadRequest, res.StatusCode)

	})

	t.Run("repository error", func(t *testing.T) {
		mockRepository := NewMockRepository(mockController)
		mockRepository.EXPECT().DeleteInvoiceById(gomock.Any(), gomock.Any()).Return(customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "repository error",
			Severity: zap.ErrorLevel,
		})

		server, validate := SetupServer(t)
		h := NewHandler(server, validate, mockRepository)
		h.RegisterRoutes()

		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("/invoices/%s", uuid.NewString()), nil)
		assert.NoError(t, err)

		res, err := server.Test(req, -1)
		assert.NoError(t, err)
		assert.Equal(t, fiber.StatusInternalServerError, res.StatusCode)
	})
}

func SetupServer(t *testing.T) (*fiber.App, *validator.Validate) {
	server := fiber.New(fiber.Config{
		JSONDecoder:           json.Unmarshal,
		JSONEncoder:           json.Marshal,
		ErrorHandler:          customError.ErrorHandler,
		DisableStartupMessage: true,
	})

	log, _ := zap.NewProduction()
	defer func(log *zap.Logger) {
		err := log.Sync()
		if err != nil {
			assert.FailNow(t, err.Error())
		}
	}(log)

	server.Use(func(c *fiber.Ctx) error {
		c.Locals("log", log)
		return c.Next()
	})

	return server, validator.New()
}
