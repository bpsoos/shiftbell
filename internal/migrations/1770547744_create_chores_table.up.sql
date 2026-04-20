CREATE TABLE chores (
    id serial primary key,
    chore_type_id integer references chore_types (id),
    name text not null,
    description text,
    last_completed_at timestamp not null,
    is_complete boolean not null,
    completed_at timestamp,
    deadline date
);
