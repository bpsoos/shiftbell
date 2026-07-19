CREATE TABLE chores (
    id integer primary key,
    chore_type_id integer references chore_types (id),
    name text not null,
    description text,
    last_completed_at timestamp,
    is_complete integer not null check (is_complete in (0, 1)),
    completed_at timestamp,
    deadline date
);
