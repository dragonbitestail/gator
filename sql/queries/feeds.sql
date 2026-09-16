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
