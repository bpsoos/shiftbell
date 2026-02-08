CREATE TABLE chores (
    id serial not null,
    description text not null,
    interval_days int not null,
    last_completed_at date not null
);
