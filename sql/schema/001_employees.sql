-- +goose Up
create table employees(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name text unique not null
);

-- +goose Down
drop table employees;