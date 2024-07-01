-- +goose Up
-- +goose StatementBegin
CREATE TABLE
    if not exists invitation (
        id integer primary key autoincrement not null,
        email text unique not null,
        token text unique not null,
        created_at DATETIME NOT NULL,
        invited_by integer not null references users (id)
    );

ALTER TABLE users
ADD COLUMN role text not null default 'user';

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROP TABLE invitation;

ALTER TABLE users
DROP COLUMN role;

-- +goose StatementEnd
