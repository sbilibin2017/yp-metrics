package repositories_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/yp-metrics/internal/repositories"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/require"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMetricDBSavePostgresContainer(t *testing.T) (*sqlx.DB, func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image: "postgres:15-alpine",
		Env: map[string]string{
			"POSTGRES_DB":       "testdb",
			"POSTGRES_USER":     "testuser",
			"POSTGRES_PASSWORD": "testpass",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForListeningPort("5432/tcp").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable", host, port.Port())

	var db *sqlx.DB
	for i := 0; i < 10; i++ {
		db, err = sqlx.ConnectContext(ctx, "pgx", dsn)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	require.NoError(t, err)

	schema := `
	CREATE SCHEMA IF NOT EXISTS content;

	CREATE TABLE IF NOT EXISTS content.metrics (
		id TEXT NOT NULL,
		mtype TEXT NOT NULL,
		delta BIGINT,
		value DOUBLE PRECISION,
		PRIMARY KEY (id, mtype)
	);
	`
	_, err = db.ExecContext(ctx, schema)
	require.NoError(t, err)

	return db, func() {
		db.Close()
		container.Terminate(ctx)
	}
}

func TestMetricDBSaveRepository_Save(t *testing.T) {
	db, cleanup := setupMetricDBSavePostgresContainer(t)
	defer cleanup()

	repo := repositories.NewMetricDBSaveRepository(db, func(ctx context.Context) *sqlx.Tx {
		return nil
	})

	ctx := context.Background()

	metric := types.Metrics{
		ID:    "metric1",
		MType: "gauge",
		Value: ptrFloat64(123.45),
	}

	err := repo.Save(ctx, metric)
	require.NoError(t, err)

	var result types.Metrics
	err = db.GetContext(ctx, &result, `
		SELECT id, mtype, delta, value FROM content.metrics WHERE id=$1 AND mtype=$2
	`, metric.ID, metric.MType)
	require.NoError(t, err)
	require.Equal(t, metric.ID, result.ID)
	require.Equal(t, metric.MType, result.MType)
	require.Nil(t, result.Delta)
	require.NotNil(t, result.Value)
	require.Equal(t, *metric.Value, *result.Value)
}

func ptrFloat64(v float64) *float64 {
	return &v
}
