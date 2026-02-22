CREATE TABLE chores (
    id serial primary key,
    chore_type_id integer references chore_types (id) not null,
    last_completed_at timestamp not null,
    completed_at timestamp,
    deadline date not null,
    is_complete boolean not null
);
