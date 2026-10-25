ALTER TABLE chores ADD COLUMN predecessor_id integer REFERENCES chores (id) ON DELETE SET NULL;

CREATE UNIQUE INDEX chores_predecessor_unique
ON chores (predecessor_id)
WHERE predecessor_id IS NOT NULL;
