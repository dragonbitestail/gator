-- +goose Up
CREATE TABLE feeds(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	name TEXT,
	url TEXT UNIQUE,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE
);


-- +goose Down
DROP TABLE feeds;
