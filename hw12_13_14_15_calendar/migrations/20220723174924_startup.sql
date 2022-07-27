-- +goose Up
-- +goose StatementBegin
create table if not exists events
(
    id varchar(40) not null
    constraint events_pk
    primary key,
    title text not null,
    description text,
    owner_id int not null,
    start timestamp not null,
    "end" timestamp not null,
    remind_before bigint not null,
    remind_sent boolean default false,
    remind_received boolean default false
    );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table if exists events;
-- +goose StatementEnd
