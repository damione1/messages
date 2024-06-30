-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages
ADD COLUMN type VARCHAR(255) NOT NULL DEFAULT 'info';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
