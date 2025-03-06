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
with inserted as (
  insert into feed_follows (id, created_at, updated_at, user_id, feed_id)
  values ( $1, $2, $3, $4, $5)
  returning *
)
SELECT inserted.*, u.name as user_name, f.name as feed_name
from inserted
join users u on inserted.user_id = u.id
join feeds f on inserted.feed_id = f.id;

-- name: GetFeedByURL :one
select * from feeds where url = $1;

-- name: GetFeedFollowsForUser :many
select feeds.name as feed_name, users.name as user_name
from feed_follows
join feeds on feed_follows.feed_id = feeds.id
join users on feed_follows.user_id = users.id
where feed_follows.user_id = $1;

-- name: DeleteFeedFollow :exec
delete from feed_follows
where user_id = $1 and feed_id = $2;