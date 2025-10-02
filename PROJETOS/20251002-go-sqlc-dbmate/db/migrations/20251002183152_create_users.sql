-- migrate:up
CREATE TABLE users (
    id uuid primary key default uuidv7(),
    username text unique
);


-- migrate:down
DROP TABLE users;

