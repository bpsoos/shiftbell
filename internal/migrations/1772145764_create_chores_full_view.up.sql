CREATE VIEW chores_full AS
    SELECT c.id,
        ct.description,
        c.last_completed_at,
        ct.interval_days,
        c.completed_at,
        c.last_completed_at + interval '1 day' * ct.interval_days as deadline,
        c.is_complete
    FROM chores c
    JOIN chore_types ct
    ON ct.id = c.chore_type_id;

