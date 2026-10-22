ALTER TABLE schedules ADD COLUMN name text NOT NULL DEFAULT '';
ALTER TABLE schedules ADD COLUMN deactivated_at timestamp;

UPDATE schedules SET name = (
    SELECT name FROM chore_templates WHERE chore_templates.id = schedules.chore_template_id
);

ALTER TABLE chores ADD COLUMN schedule_id integer REFERENCES schedules (id);
CREATE UNIQUE INDEX chores_active_schedule ON chores (schedule_id) WHERE is_complete = 0;
