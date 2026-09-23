-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
		$5,
		$6
)
RETURNING *;


-- name: GetFeed :one
SELECT * FROM feeds
	WHERE url = $1;

-- name: DeleteFeeds :exec
DELETE FROM feeds;

-- name: GetFeeds :many
SELECT * FROM feeds;

-- name: CreateFeedFollow :one
WITH i_feed_follow AS (
  INSERT INTO feed_follows(id, created_at, updated_at, user_id, feed_id)
  VALUES(
    $1,
    $2,
		$3,
    $4,
		$5
)
RETURNING *
)
SELECT u.name AS userName, f.name AS feedName, ff.*
FROM i_feed_follow ff
INNER JOIN users u ON ff.user_id = u.id
INNER JOIN feeds f ON ff.feed_id = f.id;

-- name: GetFeedFollowsForUser :many
--WITH s_followed AS (
--	SELECT u.name, f.name, ff.*
--	FROM feeds f
--	INNER JOIN feed_follows ff ON f.id = ff.feed_id
--	LEFT JOIN users u ON ff.user_id = u.id
--)
--SELECT * FROM s_followed WHERE s_followed.user_id = $1;
--SELECT * FROM (
--        SELECT u.name, f.name, ff.*
--        FROM feeds f
--        INNER JOIN feed_follows ff ON f.id = ff.feed_id
--        LEFT JOIN users u ON ff.user_id = u.id
--)
--sub_s WHERE sub_s.user_id = $1;
SELECT u.name AS userName, f.name AS feedName, f.url, ff.*
	FROM feeds f
	INNER JOIN feed_follows ff ON f.id = ff.feed_id
	LEFT JOIN users u ON ff.user_id = u.id
WHERE ff.user_id = $1;

-- name: DeleteFeedFollowForUserId :exec
DELETE FROM feed_follows
WHERE user_id = $1 AND feed_id = $2;

-- name: MarkFeedFetched :exec
UPDATE feeds
SET updated_at = CURRENT_TIMESTAMP, last_fetched_at = CURRENT_TIMESTAMP
WHERE id = $1;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at NULLS FIRST
LIMIT 1;

-- name: CreatePost :one
INSERT INTO
	posts(id, created_at, updated_at, title, url, description, published_at, feed_id)
	VALUES(
		$1,
		$2,
		$3,
		$4,
		$5,
		$6,
		$7,
		$8
	)
RETURNING *;

-- name: GetPostsForUser :many
--WITH i_feed_follow AS (
--	SELECT * FROM feed_follows
--	WHERE user_id = $1
--)
--SELECT * FROM posts p
--	WHERE i_feed_follow.feed_id = p.feed_id
--	ORDER BY published_at DESC
--	LIMIT $2;
SELECT p.id, p.created_at, p.updated_at, p.title, p.url, p.description, p.published_at, p.feed_id FROM posts p
  INNER JOIN feed_follows ff ON ff.feed_id = p.feed_id
WHERE ff.user_id = $1
ORDER BY p.published_at DESC
LIMIT $2;
