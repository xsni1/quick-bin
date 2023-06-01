CREATE TABLE files (
    id TEXT PRIMARY KEY,
    file TEXT NOT NULL,
    downloads_left SMALLINT NOT NULL,
    created_on TIMESTAMP NOT NULL
);
