-- name: CreateFeed :one
INSERT INTO feeds (name, url, user_id)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;


-- name: GetUserFeeds :many
select * from feeds where user_id = $1;

-- name: GetAllFeeds :many
select * from feeds;
