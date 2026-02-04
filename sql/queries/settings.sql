-- name: AddSet :exec
insert into settings(key,value)
values (?,?);

-- name: GetSettings :many
select * from settings;

-- name: UpdateSet :exec
update settings set value = ? where key = ?;