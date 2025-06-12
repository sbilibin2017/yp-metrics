package repositories

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/sbilibin2017/yp-metrics/internal/types"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupMetricDBGetPostgresContainer(t *testing.T) (*sqlx.DB, func()) {
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

func TestMetricDBGetRepository_Get(t *testing.T) {
	db, cleanup := setupMetricDBGetPostgresContainer(t)
	defer cleanup()

	// Prepare test data
	metricToInsert := types.Metrics{
		ID:    "metric-get-1",
		MType: "counter",
		Delta: ptrInt64(10),
		Value: nil,
	}

	_, err := db.ExecContext(context.Background(), `
		INSERT INTO content.metrics (id, mtype, delta, value) VALUES ($1, $2, $3, $4)
	`, metricToInsert.ID, metricToInsert.MType, metricToInsert.Delta, metricToInsert.Value)
	require.NoError(t, err)

	repo := NewMetricDBGetRepository(db, func(ctx context.Context) *sqlx.Tx {
		return nil
	})

	ctx := context.Background()

	gotMetric, err := repo.Get(ctx, types.MetricID{ID: metricToInsert.ID, MType: metricToInsert.MType})
	require.NoError(t, err)
	require.NotNil(t, gotMetric)
	require.Equal(t, metricToInsert.ID, gotMetric.ID)
	require.Equal(t, metricToInsert.MType, gotMetric.MType)
	require.Equal(t, metricToInsert.Delta, gotMetric.Delta)
	require.Nil(t, gotMetric.Value)
}

func ptrInt64(v int64) *int64 {
	return &v
}
