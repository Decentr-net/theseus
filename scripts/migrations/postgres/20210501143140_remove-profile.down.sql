BEGIN;

CREATE TABLE profile (
                         address TEXT PRIMARY KEY,
                         first_name TEXT NOT NULL DEFAULT '',
                         last_name TEXT NOT NULL DEFAULT '',
                         bio TEXT NOT NULL DEFAULT '',
                         avatar TEXT NOT NULL DEFAULT '',
                         gender TEXT NOT NULL DEFAULT '',
                         birthday TEXT NOT NULL DEFAULT '',
                         created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
);

COMMIT;