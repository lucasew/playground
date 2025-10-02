create table users (
    id uuid primary key default uuidv7(),
    username text unique
)
