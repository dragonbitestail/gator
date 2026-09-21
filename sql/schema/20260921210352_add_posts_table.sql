-- +goose Up
CREATE TABLE posts(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	title TEXT,
	url TEXT UNIQUE NOT NULL,
	description TEXT,
	published_at TIMESTAMP,
	feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;
