-- +goose Up
create table
	if not exists sessions (
		id integer primary key not null,
		token text not null,
		user_id integer not null references users (id),
		ip_address text,
		user_agent text,
		expires_at DATETIME not null,
		created_at DATETIME not null,
		updated_at DATETIME not null
	);

PRAGMA foreign_key_check;

-- +goose Down
drop table if exists sessions;
