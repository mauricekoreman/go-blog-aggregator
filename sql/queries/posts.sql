-- name: CreatePost :exec
insert into posts (id, created_at, updated_at, title, url, description, published_at, feed_id)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: GetPostsForUser :many
select p.title, p.url
from posts p
join feed_follows ff on p.feed_id = ff.feed_id
join users u on u.id = ff.user_id
where u.id = $1
order by p.published_at desc
limit $2;