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

func setupMetricDBListPostgresContainer(t *testing.T) (*sqlx.DB, func()) {
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

func TestMetricDBListRepository_List(t *testing.T) {
	db, cleanup := setupMetricDBListPostgresContainer(t)
	defer cleanup()

	// Insert some test data
	metricsToInsert := []types.Metrics{
		{ID: "metric1", MType: "gauge", Value: ptrFloat64(10.5)},
		{ID: "metric2", MType: "counter", Delta: ptrInt64(20)},
		{ID: "metric3", MType: "gauge", Value: ptrFloat64(30.0)},
	}

	for _, metric := range metricsToInsert {
		_, err := db.ExecContext(context.Background(), `
			INSERT INTO content.metrics (id, mtype, delta, value) VALUES ($1, $2, $3, $4)
		`, metric.ID, metric.MType, metric.Delta, metric.Value)
		require.NoError(t, err)
	}

	repo := repositories.NewMetricDBListRepository(db, func(ctx context.Context) *sqlx.Tx {
		return nil
	})

	ctx := context.Background()

	gotMetrics, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, gotMetrics, len(metricsToInsert))

	// Create maps for easy lookup
	expectedMap := map[string]types.Metrics{}
	for _, m := range metricsToInsert {
		expectedMap[m.ID] = m
	}

	for _, got := range gotMetrics {
		expected, ok := expectedMap[got.ID]
		require.True(t, ok)
		require.Equal(t, expected.ID, got.ID)
		require.Equal(t, expected.MType, got.MType)
		require.Equal(t, expected.Delta, got.Delta)
		require.Equal(t, expected.Value, got.Value)
	}
}

func ptrInt64(v int64) *int64 {
	return &v
}
