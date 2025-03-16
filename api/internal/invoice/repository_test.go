package invoice

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	customError "invoice-api/pkg/error"
)

func TestNewPgRepository(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		pgContainer := setupContainer(t)

		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		t.Cleanup(func() {
			err := pgContainer.Restore(context.Background())
			require.NoError(t, err)
		})

		assert.NotPanics(t, func() {
			NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		})
	})

	t.Run("invalid config", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPgRepository(nil, "", "", "root", "root", "test")
		})
	})

	t.Run("connection error", func(t *testing.T) {
		assert.Panics(t, func() {
			NewPgRepository(nil, "localhost", "5432", "root", "root", "test")
		})
	})
}

func TestPgRepository_CreateInvoice(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		pgContainer := setupContainer(t)
		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		t.Cleanup(func() {
			err = pgContainer.Restore(context.Background())
			require.NoError(t, err)
		})

		pgRepository := NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		err = pgRepository.CreateInvoice(context.TODO(), &InvoiceDTO{
			Id:          uuid.NewString(),
			ServiceName: "DMP",
			Amount:      120.3,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		})

		assert.NoError(t, err)
	})

	t.Run("acquire connection error", func(t *testing.T) {
		pgCfg, err := pgxpool.ParseConfig("postgres://joedoe:secret@pg.example.com:5432/mydb")
		require.NoError(t, err)

		pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
		require.NoError(t, err)

		pgRepository := &PgRepository{
			connectionPool: pool,
		}
		err = pgRepository.CreateInvoice(context.TODO(), &InvoiceDTO{
			Id:          uuid.NewString(),
			ServiceName: "DMP",
			Amount:      120.3,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		})

		assert.Error(t, err)
	})

	t.Run("repository error", func(t *testing.T) {
		pgContainer := setupContainer(t)
		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		t.Cleanup(func() {
			err = pgContainer.Restore(context.Background())
			require.NoError(t, err)
		})

		pgRepository := NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		err = pgRepository.CreateInvoice(context.TODO(), &InvoiceDTO{
			Id:          uuid.NewString(),
			ServiceName: "DMP",
			Amount:      -1,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		})

		assert.NoError(t, err)
	})
}

func TestPgRepository_UpdateInvoice(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		pgContainer := setupContainer(t)
		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		t.Cleanup(func() {
			err = pgContainer.Restore(context.Background())
			require.NoError(t, err)
		})

		invoiceId := uuid.NewString()
		pgRepository := NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		_, err = pgRepository.connectionPool.Exec(
			context.TODO(),
			"insert into invoices (id, service_name, amount, status, date) values ($1, $2, $3, $4, $5)",
			invoiceId,
			"DMP",
			120.3,
			"PAID",
			time.Now().UTC(),
		)
		require.NoError(t, err)

		err = pgRepository.UpdateInvoiceById(context.TODO(), invoiceId, &InvoiceDTO{
			Id:          invoiceId,
			ServiceName: "DMP",
			Amount:      120.3,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		})
		assert.NoError(t, err)
	})

	t.Run("acquire connection error", func(t *testing.T) {
		pgCfg, err := pgxpool.ParseConfig("postgres://joedoe:secret@pg.example.com:5432/mydb")
		require.NoError(t, err)

		pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
		require.NoError(t, err)

		pgRepository := &PgRepository{
			connectionPool: pool,
		}
		err = pgRepository.UpdateInvoiceById(context.TODO(), uuid.NewString(), &InvoiceDTO{
			Id:          uuid.NewString(),
			ServiceName: "DMP",
			Amount:      120.3,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		})

		assert.Error(t, err)
	})

	t.Run("not found", func(t *testing.T) {
		pgContainer := setupContainer(t)
		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		t.Cleanup(func() {
			err = pgContainer.Restore(context.Background())
			require.NoError(t, err)
		})

		invoiceId := uuid.NewString()
		pgRepository := NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		err = pgRepository.UpdateInvoiceById(context.TODO(), invoiceId, &InvoiceDTO{
			Id:          invoiceId,
			ServiceName: "DMP",
			Amount:      120.3,
			Status:      "PAID",
			Date:        time.Now().UTC(),
		})
		assert.NoError(t, err)
	})
}

func TestPgRepository_GetInvoiceById(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		pgContainer := setupContainer(t)
		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		t.Cleanup(func() {
			err = pgContainer.Restore(context.Background())
			require.NoError(t, err)
		})

		invoiceId := uuid.NewString()
		pgRepository := NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		_, err = pgRepository.connectionPool.Exec(
			context.TODO(),
			"insert into invoices (id, service_name, amount, status, date) values ($1, $2, $3, $4, $5)",
			invoiceId,
			"DMP",
			120.3,
			"PAID",
			time.Now().UTC(),
		)
		require.NoError(t, err)

		invoice, err := pgRepository.GetInvoiceById(context.TODO(), invoiceId)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, invoiceId, invoice.Id)
	})

	t.Run("acquire connection error", func(t *testing.T) {
		pgCfg, err := pgxpool.ParseConfig("postgres://joedoe:secret@pg.example.com:5432/mydb")
		require.NoError(t, err)

		pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
		require.NoError(t, err)

		pgRepository := &PgRepository{
			connectionPool: pool,
		}
		invoice, err := pgRepository.GetInvoiceById(context.TODO(), uuid.NewString())

		assert.Error(t, err)
		assert.Nil(t, invoice)
	})

	t.Run("not found", func(t *testing.T) {
		pgContainer := setupContainer(t)
		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		pgRepository := NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		invoice, err := pgRepository.GetInvoiceById(context.TODO(), uuid.NewString())

		assert.Nil(t, invoice)
		assert.Error(t, err)
		assert.Equal(t, err.(customError.CustomError).Code, http.StatusNotFound)
	})
}

func TestPgRepository_DeleteInvoiceById(t *testing.T) {
	t.Run("happy path", func(t *testing.T) {
		pgContainer := setupContainer(t)
		pgHost, err := pgContainer.Host(context.Background())
		require.NoError(t, err)

		pgPort, err := pgContainer.MappedPort(context.Background(), "5432/tcp")
		require.NoError(t, err)

		t.Cleanup(func() {
			err = pgContainer.Restore(context.Background())
			require.NoError(t, err)
		})

		invoiceId := uuid.NewString()
		pgRepository := NewPgRepository(nil, pgHost, pgPort.Port(), "root", "root", "test")
		_, err = pgRepository.connectionPool.Exec(
			context.TODO(),
			"insert into invoices (id, service_name, amount, status, date) values ($1, $2, $3, $4, $5)",
			invoiceId,
			"DMP",
			120.3,
			"PAID",
			time.Now().UTC(),
		)
		require.NoError(t, err)

		err = pgRepository.DeleteInvoiceById(context.TODO(), invoiceId)

		assert.NoError(t, err)
	})

	t.Run("acquire connection error", func(t *testing.T) {
		pgCfg, err := pgxpool.ParseConfig("postgres://joedoe:secret@pg.example.com:5432/mydb")
		require.NoError(t, err)

		pool, err := pgxpool.NewWithConfig(context.Background(), pgCfg)
		require.NoError(t, err)

		pgRepository := &PgRepository{
			connectionPool: pool,
		}
		err = pgRepository.DeleteInvoiceById(context.TODO(), uuid.NewString())

		assert.NoError(t, err)
	})
}

func setupContainer(t *testing.T) *postgres.PostgresContainer {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(
		ctx,
		"postgres",
		postgres.WithDatabase("test"),
		postgres.WithUsername("root"),
		postgres.WithPassword("root"),
		postgres.BasicWaitStrategies(),
		postgres.WithSQLDriver("pgx"),
		postgres.WithInitScripts("../../.scripts/init.sql"),
	)
	require.NoError(t, err)

	t.Cleanup(func() {
		err = postgresContainer.Terminate(ctx)
		require.NoError(t, err)
	})

	err = postgresContainer.Snapshot(ctx)
	require.NoError(t, err)

	return postgresContainer
}
