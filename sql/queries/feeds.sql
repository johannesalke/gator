-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at,  name, url, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;


-- name: GetUserFeeds :many
select * from feeds where user_id = $1;

-- name: GetAllFeeds :many
select * from feeds;

-- name: GetFeedByURL :one
select * from feeds where url = $1;

-- name: MarkFeedFetched :one
update feeds set updated_at = NOW(), last_fetched_at = NOW() where id = $1
returning *;


-- name: GetNextFeedToFetch :one

SELECT * FROM feeds ORDER BY last_fetched_at asc nulls first limit 1;


