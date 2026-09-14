-- +goose Up
CREATE TABLE users(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	name TEXT UNIQUE
);

-- +goose Down
DROP TABLE users;
