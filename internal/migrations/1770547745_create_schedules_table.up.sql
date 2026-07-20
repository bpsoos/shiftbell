CREATE TABLE schedules (
    id integer primary key,
    chore_template_id integer references chore_templates (id) not null,
    interval_days int not null
);
