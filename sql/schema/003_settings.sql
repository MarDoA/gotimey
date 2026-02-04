-- +goose Up
create table settings(
    key text primary key,
    value text not null
);

-- +goose Down
drop table settings;