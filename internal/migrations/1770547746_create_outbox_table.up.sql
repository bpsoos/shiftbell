CREATE TABLE outbox (
    id serial primary key,
    status text not null,
    type text not null,
    payload jsonb
);
