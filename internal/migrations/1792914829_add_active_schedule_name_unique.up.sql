CREATE UNIQUE INDEX schedules_active_name_unique
ON schedules (lower(name))
WHERE deactivated_at IS NULL;
