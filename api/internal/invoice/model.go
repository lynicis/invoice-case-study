package invoice

import (
	"time"
)

type CreateInvoiceRequest struct {
	ServiceName string    `json:"serviceName" validate:"required,oneof=DMP SSP"`
	Amount      float32   `json:"amount" validate:"required,min=1"`
	Status      string    `json:"status" validate:"required,oneof=PAID UNPAID PENDING"`
	Date        time.Time `json:"date" validate:"required"`
}

type UpdateInvoiceRequest struct {
	CreateInvoiceRequest
	Id string `json:"id" validate:"required,uuid4"`
}

type GetInvoicesRequest struct {
	Page     int    `query:"page,omitempty"`
	PageSize int    `query:"pageSize,omitempty"`
	Search   string `query:"search,omitempty"`
}

type InvoiceDTO struct {
	Id          string    `json:"id" db:"id"`
	ServiceName string    `json:"serviceName" db:"service_name"`
	Amount      float32   `json:"amount" db:"amount"`
	Status      string    `json:"status" db:"status"`
	Date        time.Time `json:"date" db:"date"`
}
