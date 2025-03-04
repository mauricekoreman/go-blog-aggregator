-- name: AddFeed :one
insert into feeds (id, created_at, updated_at, name, url, user_id)
values ( $1, $2, $3, $4, $5, $6)
returning *;

-- name: GetAllFeeds :many
select f.name, f.url, u.name as user_name
from feeds f 
join users u 
on f.user_id = u.id;

-- name: CreateFeedFollow :one
insert into feed_follows (id, created_at, updated_at, user_id, feed_id)
values ( $1, $2, $3, $4, $5)