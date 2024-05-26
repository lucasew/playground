create table if not exists blobs(
    description text default "",
    data blob not null
)
