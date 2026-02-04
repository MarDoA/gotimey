-- +goose Up
create table attendance(
    id INTEGER primary key AUTOINCREMENT,
    start INTEGER not null,
    end INTEGER,
    hours INTEGER,
    emp_id int not null,
    foreign key (emp_id) references employees(id)
);

-- +goose Down
drop table attendance;