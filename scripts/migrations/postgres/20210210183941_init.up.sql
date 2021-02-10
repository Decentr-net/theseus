BEGIN;

CREATE TABLE height (
    height BIGINT
);

INSERT INTO height VALUES(0);

CREATE TABLE post (
    owner TEXT NOT NULL,
    uuid TEXT NOT NULL,
    title TEXT NOT NULL,
    category INT8 NOT NULL,
    preview_image TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by TEXT,

    PRIMARY KEY (owner, uuid)
);

CREATE INDEX post_created_at_idx ON post(created_at DESC);
CREATE INDEX post_category_idx ON post(category);

CREATE TABLE "like" (
    post_owner TEXT NOT NULL,
    post_uuid TEXT NOT NULL,
    liked_by TEXT NOT NULL,
    weight INT2 NOT NULL,
    liked_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    PRIMARY KEY (post_owner, post_uuid, liked_by),
    FOREIGN KEY (post_owner, post_uuid) REFERENCES post(owner, uuid)
);

COMMIT;