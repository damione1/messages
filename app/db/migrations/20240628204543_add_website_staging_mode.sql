-- +goose Up
-- +goose StatementBegin
ALTER TABLE websites
ADD COLUMN staging BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
