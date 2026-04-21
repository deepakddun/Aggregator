-- name: CreateUser :one
INSERT INTO users (id, created_at, name)
VALUES (
    $1,
    $2,
    $3
)
RETURNING *;



-- name: GetUser :one
SELECT * FROM users
WHERE name = $1;

-- name: DeleteUser :exec
DELETE FROM users;

-- name: GetUsers :many
SELECT name FROM users
ORDER BY name;

-- name: CreateFeed :one
INSERT INTO feeds (id, name , url , created_at, user_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5
)
RETURNING *;


-- name: ListFeeds :many
SELECT f.name , f.url , u.name FROM users u JOIN feeds f
ON u.id = f.user_id;

-- name: CreateFeedFollow :many
WITH inserted_feed_follows AS (
    INSERT INTO feed_follows (id , created_at,user_id , feed_id) 
    VALUES (
        $1,
        $2,
        $3,
        $4
    )
    RETURNING *
)
SELECT 
    inserted_feed_follows.*,
    users.name AS USER_NAME,
    feeds.name AS FEED_NAME
FROM
    inserted_feed_follows,
    users,
    feeds
WHERE    
   inserted_feed_follows.user_id = users.id
   AND inserted_feed_follows.feed_id = feeds.id;


-- name: GetFeedByURL :one
SELECT * FROM feeds
WHERE url = $1;

-- name: GetFeedFollowsForUser :many
SELECT f.name as feed_name , u.name as user_name FROM 
feed_follows ff ,
feeds f,
users u
where ff.feed_id = f.id
and ff.user_id = u.id
and u.id = $1;


-- name: DeleteFeedFollow :exec
DELETE FROM feed_follows
WHERE user_id=$1 AND feed_id=$2;


-- name: MarkFeedFetched :exec
UPDATE feeds
SET updated_at=now() , last_fetched_at=now()
where id = $1;

-- name: GetNextFeedToFetch :one
select * from feeds f 
order by last_fetched_at asc nulls first
limit 1;

-- name: CreatePost :one
INSERT INTO POSTS (id, url , title , description , created_at, published_at,feed_id)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6,
    $7
)
RETURNING *;



-- name: GetPostsByUser :many
select ff.* from feeds f , feed_follows ff
where f.id = ff.feed_id
and ff.user_id = $1
order by f.created_at desc , f.updated_at desc
LIMIT $2;