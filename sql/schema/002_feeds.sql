-- +goose Up
CREATE TABLE feeds(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	name TEXT,
	url TEXT UNIQUE,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE feed_follows(
	id UUID PRIMARY KEY,
	created_at TIMESTAMP,
	updated_at TIMESTAMP,
	user_id UUID REFERENCES users(id) ON DELETE CASCADE,
	feed_id UUID REFERENCES feeds(id) ON DELETE CASCADE,
	UNIQUE(user_id, feed_id)
);

-- +goose Down
DROP TABLE feeds;
DROP TABLE feed_follows;
