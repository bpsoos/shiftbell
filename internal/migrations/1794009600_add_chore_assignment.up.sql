ALTER TABLE chores ADD COLUMN owner_profile_id integer REFERENCES profiles(id);
