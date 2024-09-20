CREATE TABLE IF NOT EXISTS users(
    guid UUID DEFAULT gen_random_uuid(),
    email TEXT NOT NULL,
    password TEXT NOT NULL,
    reftokenhash BYTEA NOT NULL,
    PRIMARY KEY(guid)
);
