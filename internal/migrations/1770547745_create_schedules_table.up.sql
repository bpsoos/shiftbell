CREATE TABLE schedules (
    id serial primary key,
    chore_type_id integer references chore_types (id) not null,
    interval_days int not null
);
