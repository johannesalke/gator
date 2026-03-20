-- name: CreateFeedFollowSimple :one
INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
    
) 
RETURNING *;


-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
    INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
        $1,
        $2,
        $3,
        $4,
        $5
       
    )
    RETURNING *
) 
SELECT
    inserted_feed_follow.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM inserted_feed_follow
INNER JOIN feeds ON feeds.id = inserted_feed_follow.feed_id
INNER JOIN users ON users.id = inserted_feed_follow.user_id
;

-- name: GetFeedFollowsForUser :many
WITH retrieve_feed_follows AS (
Select * from feed_follows
where feed_follows.user_id = $1

)
SELECT
    retrieve_feed_follows.*,
    feeds.name AS feed_name,
    users.name AS user_name
FROM retrieve_feed_follows
INNER JOIN feeds ON feeds.id = retrieve_feed_follows.feed_id
INNER JOIN users ON users.id = retrieve_feed_follows.user_id
;

-- name: UnfollowFeed :one
Delete from feed_follows
where feed_id = $1 and user_id = $2
RETURNING *;