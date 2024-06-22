-- +goose Up
PRAGMA foreign_keys = ON;

create table
	if not exists users (
		id integer primary key not null,
		email text unique not null,
		password_hash text not null,
		first_name text not null,
		last_name text not null,
		email_verified_at DATETIME,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);

-- +goose Down
drop table if exists users;
