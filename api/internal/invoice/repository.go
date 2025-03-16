package invoice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	customError "invoice-api/pkg/error"
)

type Repository interface {
	CreateInvoice(ctx context.Context, invoice *InvoiceDTO) error
	GetInvoices(ctx context.Context, page int, pageSize int, search string) (*[]InvoiceDTO, error)
	GetInvoiceById(ctx context.Context, id string) (*InvoiceDTO, error)
	UpdateInvoiceById(ctx context.Context, id string, invoice *InvoiceDTO) error
	DeleteInvoiceById(ctx context.Context, id string) error
}

type PgRepository struct {
	connectionPool *pgxpool.Pool
}

func NewPgRepository(log *zap.Logger, host, port, username, password, database string) *PgRepository {
	credentials := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s sslmode=disable", username, password, host, port, database)
	pgConfig, err := pgxpool.ParseConfig(credentials)
	if err != nil {
		log.Fatal("failed to parse database config", zap.Error(err))
	}

	var pgConnectionPool *pgxpool.Pool
	pgConnectionPool, err = pgxpool.NewWithConfig(context.Background(), pgConfig)
	if err != nil {
		log.Fatal("failed to connect database", zap.Error(err))
	}

	var connection *pgxpool.Conn
	connection, err = pgConnectionPool.Acquire(context.Background())
	if err != nil {
		log.Fatal("failed to acquire connection", zap.Error(err))
	}
	defer connection.Release()

	err = connection.Ping(context.Background())
	if err != nil {
		log.Fatal("failed to ping database", zap.Error(err))
	}

	return &PgRepository{
		connectionPool: pgConnectionPool,
	}
}

func (r *PgRepository) CreateInvoice(ctx context.Context, invoice *InvoiceDTO) error {
	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to acquire connection",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}
	defer connection.Release()

	if _, err = connection.Exec(
		ctx,
		"insert into invoices (id, service_name, amount, status, date) values ($1, $2, $3, $4, $5)",
		invoice.Id,
		invoice.ServiceName,
		invoice.Amount,
		invoice.Status,
		time.Now().UTC(),
	); err != nil {
		return customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to create invoice",
			Severity: zap.WarnLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	return nil
}

func (r *PgRepository) GetInvoices(
	ctx context.Context,
	page,
	pageSize int,
	search string,
) (*[]InvoiceDTO, error) {
	var query strings.Builder
	query.WriteString("select * from invoices")

	args := make([]interface{}, 0, 3)
	argIndex := 1

	if search != "" {
		query.WriteString(" where to_tsvector(id || ' ' || service_name) @@ to_tsquery($1)")
		args = append(args, search)
		argIndex++
	}

	query.WriteString(fmt.Sprintf(" limit $%d offset $%d", argIndex, argIndex+1))
	args = append(args, pageSize, (page-1)*pageSize)

	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to acquire connection",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}
	defer connection.Release()

	var rows pgx.Rows
	rows, err = connection.Query(ctx, query.String(), args...)
	if err != nil {
		return nil, customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to get invoices",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	var invoices []InvoiceDTO
	invoices, err = pgx.CollectRows(rows, pgx.RowToStructByName[InvoiceDTO])
	if err != nil {
		return nil, customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to collect invoices",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	return &invoices, nil
}

func (r *PgRepository) GetInvoiceById(ctx context.Context, id string) (*InvoiceDTO, error) {
	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return nil, customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to acquire connection",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}
	defer connection.Release()

	var row pgx.Rows
	row, err = connection.Query(ctx, "select * from invoices where id = $1", id)
	if err != nil {
		return nil, customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to get invoice",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	var invoice InvoiceDTO
	invoice, err = pgx.CollectOneRow(row, pgx.RowToStructByPos[InvoiceDTO])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customError.CustomError{
				Code:     fiber.StatusNotFound,
				Message:  "invoice not found",
				Severity: zap.WarnLevel,
			}
		}

		return nil, customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to collect an invoice",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	return &invoice, nil
}

func (r *PgRepository) UpdateInvoiceById(ctx context.Context, id string, invoice *InvoiceDTO) error {
	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to acquire connection",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}
	defer connection.Release()

	if _, err = connection.Exec(
		ctx,
		"update invoices set service_name = $1, amount = $2, status = $3 where id = $4",
		invoice.ServiceName,
		invoice.Amount,
		invoice.Status,
		id,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customError.CustomError{
				Code:     fiber.StatusNotFound,
				Message:  "invoice not found",
				Severity: zap.WarnLevel,
			}
		}
		return customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to update invoice by id",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	return nil
}

func (r *PgRepository) DeleteInvoiceById(ctx context.Context, id string) error {
	connection, err := r.connectionPool.Acquire(ctx)
	if err != nil {
		return customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to acquire connection",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}
	defer connection.Release()

	if _, err = connection.Exec(ctx, "delete from invoices where id = $1", id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return customError.CustomError{
				Code:     fiber.StatusNotFound,
				Message:  "invoice not found",
				Severity: zap.WarnLevel,
			}
		}

		return customError.CustomError{
			Code:     fiber.StatusInternalServerError,
			Message:  "failed to delete an invoice",
			Severity: zap.ErrorLevel,
			Fields:   []zap.Field{zap.Error(err)},
		}
	}

	return nil
}
