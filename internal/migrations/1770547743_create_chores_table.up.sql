CREATE TABLE chore_types (
    id serial primary key,
    description text not null,
    interval_days int not null
);
