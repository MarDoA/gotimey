-- name: CreateAttendanceStart :one
insert into attendance (start,emp_id)
values (?,?)
RETURNING id,start;

-- name: GetActiveSessionByEmpID :one
select id from attendance where emp_id = ? and end is null order by start desc limit 1;

-- name: AddEndToAttendaceByID :one
update attendance set end = ?,hours =?-start where id = ?
RETURNING start,end,hours;

-- name: GetAttendanceByMonth :many
select * from attendance where start > ? and start < ? and emp_id =?;

-- name: GetAttendanceByID :one
select * from attendance where id=?;

-- name: UpdateAttendaceByID :exec
update attendance set start = ?, end =?, hours=?,emp_id=? where id=?;

-- name: DeleteEntry :exec
delete from attendance where id =?;