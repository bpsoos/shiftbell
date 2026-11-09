CREATE TABLE accounts (
    id integer primary key AUTOINCREMENT,
    username text NOT NULL,
    password_hash text NOT NULL
);

CREATE UNIQUE INDEX accounts_single_account
ON accounts ((1));
