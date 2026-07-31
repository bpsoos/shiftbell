CREATE TABLE chore_templates (
    id integer primary key,
    name text not null,
    description text,
    deactivated_at timestamp
);
