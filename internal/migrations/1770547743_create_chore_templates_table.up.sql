CREATE TABLE chore_templates (
    id integer primary key,
    name text not null,
    description text,
    deactivated_at timestamp
);

CREATE UNIQUE INDEX chore_templates_active_name_unique
ON chore_templates (lower(name))
WHERE deactivated_at IS NULL;
