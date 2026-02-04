-- name: CreateEmployee :one
insert into employees (name)
values (?)
RETURNING id;

-- name: GetEmployees :many
select * from employees;

-- name: DeleteEmployee :exec
delete from employees where id =?;