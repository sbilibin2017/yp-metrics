-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS content;

CREATE TABLE content.metrics (
    id TEXT NOT NULL,
    mtype TEXT NOT NULL,
    delta BIGINT,
    value DOUBLE PRECISION,
    PRIMARY KEY (id, mtype)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS content.metrics;
DROP SCHEMA IF EXISTS content;
-- +goose StatementEnd
