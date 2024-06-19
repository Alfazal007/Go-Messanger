-- name: CreateUser :one
insert into users (id, name, email, password, created_at, updated_at) values ($1, $2, $3, $4, $5, $6) returning *;

-- name: GetUserById :one
select * from users where id=$1;

-- name: GetUserByEmail :one
select * from users where email=$1;

-- name: GetUserBeforeCreate :one
select count(*) from users where name=$1 or email=$2;
