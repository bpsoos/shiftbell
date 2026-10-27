BEGIN IMMEDIATE;

INSERT INTO chore_templates (id, name, description) VALUES
    (1, 'Clean kitchen counters', 'Clear and wipe every work surface.'),
    (2, 'Take out rubbish', 'Empty household bins and replace liners.'),
    (3, 'Water houseplants', 'Check the soil before watering each plant.'),
    (4, 'Vacuum living areas', 'Vacuum rugs, floors, and beneath loose furniture.'),
    (5, 'Mop hard floors', 'Mop the kitchen, hall, and bathroom floors.'),
    (6, 'Clean bathroom', 'Clean the sink, toilet, bath, and mirrors.'),
    (7, 'Change bed linen', 'Replace sheets, pillowcases, and duvet cover.'),
    (8, 'Dust shelves', 'Dust shelves, frames, lamps, and electronics.'),
    (9, 'Check fridge', 'Discard expired food and note anything running low.'),
    (10, 'Deep clean oven', 'Clean the racks, trays, door, and interior.'),
    (11, 'Wash windows', 'Clean the inside glass and window sills.'),
    (12, 'Descale kettle', 'Descale, rinse thoroughly, and refill.'),
    (13, 'Rotate mattress', 'Rotate the mattress from head to foot.'),
    (14, 'Test smoke alarms', 'Test every alarm and replace weak batteries.'),
    (15, 'Clean gutters', 'Remove leaves and confirm the downpipes drain.'),
    (16, 'Organize pantry', 'Group similar food and move older items forward.'),
    (17, 'Wash towels', 'Wash bath towels, hand towels, and bath mats.'),
    (18, 'Clean coffee machine', 'Wash removable parts and run a cleaning cycle.'),
    (19, 'Wipe cabinet fronts', 'Wipe handles and the outside of cabinet doors.'),
    (20, 'Sort recycling', 'Separate recycling and take full containers out.'),
    (21, 'Clean entryway', 'Sweep the floor and put away shoes and bags.'),
    (22, 'Refresh pet bedding', 'Wash bedding and clean the surrounding area.'),
    (23, 'Back up photos', 'Copy recent photos to the household backup.'),
    (24, 'Review first-aid kit', 'Replace expired or missing supplies.');

INSERT INTO chores (id, name, description, is_complete, completed_on, deadline) VALUES
    (1, 'Clean kitchen counters', 'Clear and wipe every work surface.', 0, NULL, date('now', 'localtime', '-10 days')),
    (2, 'Take out rubbish', 'Empty household bins and replace liners.', 0, NULL, date('now', 'localtime', '-4 days')),
    (3, 'Water houseplants', 'Check the soil before watering each plant.', 0, NULL, date('now', 'localtime', '-1 day')),
    (4, 'Feed sourdough starter', 'Refresh the starter and leave it at room temperature.', 0, NULL, date('now', 'localtime')),
    (5, 'Wipe bathroom sink', NULL, 0, NULL, date('now', 'localtime')),
    (6, 'Replace kitchen towels', 'Put out clean hand towels and tea towels.', 0, NULL, date('now', 'localtime', '+1 day')),
    (7, 'Vacuum living areas', 'Vacuum rugs, floors, and beneath loose furniture.', 0, NULL, date('now', 'localtime', '+2 days')),
    (8, 'Mop hard floors', 'Mop the kitchen, hall, and bathroom floors.', 0, NULL, date('now', 'localtime', '+4 days')),
    (9, 'Clean bathroom', 'Clean the sink, toilet, bath, and mirrors.', 0, NULL, date('now', 'localtime', '+6 days')),
    (10, 'Change bed linen', 'Replace sheets, pillowcases, and duvet cover.', 0, NULL, date('now', 'localtime', '+8 days')),
    (11, 'Dust shelves', 'Dust shelves, frames, lamps, and electronics.', 0, NULL, date('now', 'localtime', '+10 days')),
    (12, 'Check fridge', 'Discard expired food and note anything running low.', 0, NULL, date('now', 'localtime', '+13 days')),
    (13, 'Deep clean oven', 'Clean the racks, trays, door, and interior.', 0, NULL, date('now', 'localtime', '+14 days')),
    (14, 'Wash windows', 'Clean the inside glass and window sills.', 0, NULL, date('now', 'localtime', '+20 days')),
    (15, 'Descale kettle', 'Descale, rinse thoroughly, and refill.', 0, NULL, date('now', 'localtime', '+21 days')),
    (16, 'Rotate mattress', 'Rotate the mattress from head to foot.', 0, NULL, date('now', 'localtime', '+27 days')),
    (17, 'Test smoke alarms', 'Test every alarm and replace weak batteries.', 0, NULL, date('now', 'localtime', '+28 days')),
    (18, 'Clean gutters', 'Remove leaves and confirm the downpipes drain.', 0, NULL, date('now', 'localtime', '+35 days')),
    (19, 'Organize pantry', 'Group similar food and move older items forward.', 0, NULL, date('now', 'localtime', '+42 days')),
    (20, 'Wash towels', 'Wash bath towels, hand towels, and bath mats.', 0, NULL, date('now', 'localtime', '+49 days')),
    (21, 'Clean coffee machine', 'Wash removable parts and run a cleaning cycle.', 0, NULL, date('now', 'localtime', '+60 days')),
    (22, 'Wipe cabinet fronts', 'Wipe handles and the outside of cabinet doors.', 0, NULL, date('now', 'localtime', '+75 days')),
    (23, 'Sort recycling', NULL, 0, NULL, date('now', 'localtime', '+90 days')),
    (24, 'Clean entryway', 'Sweep the floor and put away shoes and bags.', 0, NULL, date('now', 'localtime', '+120 days')),
    (25, 'Refresh pet bedding', 'Wash bedding and clean the surrounding area.', 0, NULL, date('now', 'localtime', '+150 days')),
    (26, 'Back up photos', 'Copy recent photos to the household backup.', 0, NULL, date('now', 'localtime', '+180 days')),
    (27, 'Review first-aid kit', 'Replace expired or missing supplies.', 0, NULL, date('now', 'localtime', '+210 days')),
    (28, 'Clean balcony', 'Sweep the floor and wipe the outdoor furniture.', 0, NULL, date('now', 'localtime', '+240 days')),
    (29, 'Defrost freezer', 'Use up loose food before defrosting and cleaning.', 0, NULL, date('now', 'localtime', '+270 days')),
    (30, 'Donate unused clothes', 'Fill one bag with clean items ready to donate.', 0, NULL, date('now', 'localtime', '+300 days')),
    (31, 'Clean kitchen counters', 'Clear and wipe every work surface.', 1, datetime('now', 'localtime', '-2 days'), date('now', 'localtime', '-2 days')),
    (32, 'Take out rubbish', 'Empty household bins and replace liners.', 1, datetime('now', 'localtime', '-5 days'), date('now', 'localtime', '-5 days')),
    (33, 'Vacuum bedrooms', NULL, 1, datetime('now', 'localtime', '-8 days'), date('now', 'localtime', '-8 days')),
    (34, 'Wash towels', 'Wash bath towels, hand towels, and bath mats.', 1, datetime('now', 'localtime', '-12 days'), date('now', 'localtime', '-12 days')),
    (35, 'Clean coffee machine', 'Wash removable parts and run a cleaning cycle.', 1, datetime('now', 'localtime', '-18 days'), date('now', 'localtime', '-18 days')),
    (36, 'Change bed linen', 'Replace sheets, pillowcases, and duvet cover.', 1, datetime('now', 'localtime', '-25 days'), date('now', 'localtime', '-25 days'));

INSERT INTO schedules (id, name, chore_template_id, interval_days) VALUES
    (1, 'Weekly plant care', 3, 7),
    (2, 'Daily kitchen reset', 1, 1),
    (3, 'Rubbish collection', 2, 3),
    (4, 'Weekly vacuuming', 4, 7),
    (5, 'Weekly floor cleaning', 5, 7),
    (6, 'Bathroom cleaning', 6, 7),
    (7, 'Fresh bed linen', 7, 14),
    (8, 'Shelf dusting', 8, 14),
    (9, 'Fridge check', 9, 7),
    (10, 'Oven cleaning', 10, 90),
    (11, 'Window cleaning', 11, 90),
    (12, 'Kettle descaling', 12, 30),
    (13, 'Mattress rotation', 13, 90),
    (14, 'Smoke alarm check', 14, 30),
    (15, 'Gutter maintenance', 15, 180),
    (16, 'Pantry tidy', 16, 30),
    (17, 'Towel laundry', 17, 7),
    (18, 'Coffee machine cleaning', 18, 7),
    (19, 'Cabinet cleaning', 19, 14),
    (20, 'Recycling collection', 20, 7),
    (21, 'Entryway reset', 21, 7);

INSERT INTO chores (id, name, description, is_complete, deadline, schedule_id)
SELECT 37, name, description, 0, date('now', 'localtime', '+1 day'), 1
FROM chore_templates WHERE id = 3;

INSERT INTO chores (id, name, description, is_complete, deadline, schedule_id)
SELECT 36 + s.id, t.name, t.description, 0,
    date('now', 'localtime', printf('+%d days', s.id - 2)), s.id
FROM schedules s
JOIN chore_templates t ON t.id = s.chore_template_id
WHERE s.id BETWEEN 2 AND 21;

INSERT INTO schedules (id, name, chore_template_id, interval_days, deactivated_at) VALUES
    (22, 'Previous window cleaning', 11, 90, datetime('now', 'localtime', '-30 days')),
    (23, 'Previous gutter maintenance', 15, 180, datetime('now', 'localtime', '-14 days'));

COMMIT;
