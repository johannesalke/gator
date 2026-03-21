-- name: CreatePost :one
INSERT INTO posts (id, created_at, updated_at,  title, url, description, published_at, feed_id)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING *;

-- name: GetPostsForUser :many
WITH user_feeds AS (
    Select feed_id from feed_follows 
    where user_id = $1
) 
select * from posts 
where feed_id in (SELECT feed_id FROM user_feeds) 
ORDER by published_at desc nulls last limit $2;

