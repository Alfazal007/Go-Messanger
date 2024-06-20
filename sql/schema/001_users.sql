-- +goose Up

create table users (
    id uuid primary key,
    username text not null unique,
    password text not null,
    email text not null unique,
    refresh_token text,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL
);

-- +goose Down
drop table users;
