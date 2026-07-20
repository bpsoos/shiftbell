CREATE TABLE chores (
    id integer primary key,
    chore_template_id integer references chore_templates (id),
    name text not null,
    description text,
    is_complete integer not null check (is_complete in (0, 1)),
    completed_on timestamp,
    deadline date
);
