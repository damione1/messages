-- +goose Up
-- +goose StatementBegin
ALTER TABLE websites
RENAME COLUMN websiteName TO name;

ALTER TABLE websites
RENAME COLUMN websiteUrl TO url;

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
