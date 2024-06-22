-- +goose Up
-- +goose StatementBegin
-- Create the websites table
create table
    if not exists websites (
        id integer primary key autoincrement not null,
        websiteName text not null,
        websiteUrl text not null
    );

-- Create the messages table
create table
    if not exists messages (
        id integer primary key autoincrement not null,
        title text not null,
        message text not null,
        language text not null,
        userId integer not null references users (id),
        display_from DATETIME NOT NULL,
        display_to DATETIME NOT NULL,
        created_at DATETIME NOT NULL,
        updated_at DATETIME NOT NULL
    );

-- Create the website_messages table (join table)
create table
    if not exists websites_messages (
        id integer primary key autoincrement not null,
        websiteId integer not null references websites (id),
        messageId integer not null references messages (id)
    );

-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
SELECT
    'down SQL query';

-- +goose StatementEnd
