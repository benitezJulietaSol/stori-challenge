CREATE DATABASE postgres;

CREATE USER postgres WITH PASSWORD 'admin';

create table transactions
(
    id         integer                 not null,
    amount     integer                 not null,
    date       varchar                 not null,
    created_at timestamp default now() not null,
    primary key (id, created_at)
);

alter table transactions
    owner to postgres;