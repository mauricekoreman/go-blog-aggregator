-- +goose Up
create table posts (
  id UUID primary key,
  created_at timestamp not null,
  updated_at timestamp not null,
  title text not null,
  url text not null unique,
  description text,
  published_at timestamp,
  feed_id UUID not null references feeds(id)
);

-- +goose Down
drop table posts;