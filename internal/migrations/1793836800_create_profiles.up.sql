CREATE TABLE profiles (
    id integer primary key,
    name text NOT NULL,
    name_key text NOT NULL UNIQUE
);
