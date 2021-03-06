BEGIN;

CREATE TABLE height (
    height BIGINT
);

INSERT INTO height VALUES(0);

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

CREATE TABLE updv (
    address TEXT REFERENCES profile(address),
    updv BIGINT NOT NULL,
    timestamp TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    PRIMARY KEY (address, timestamp)
);

CREATE TABLE follow (
    follower TEXT REFERENCES profile(address),
    followee TEXT REFERENCES profile(address),

    PRIMARY KEY (follower, followee)
);

CREATE TABLE post (
    owner TEXT NOT NULL REFERENCES profile(address),
    uuid TEXT NOT NULL,
    title TEXT NOT NULL,
    category INT8 NOT NULL,
    preview_image TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,
    deleted_at TIMESTAMP WITHOUT TIME ZONE,
    deleted_by TEXT REFERENCES profile(address),

    PRIMARY KEY (owner, uuid)
);

CREATE TABLE "like" (
    post_owner TEXT NOT NULL REFERENCES profile(address),
    post_uuid TEXT NOT NULL,
    liked_by TEXT NOT NULL REFERENCES profile(address),
    weight INT2 NOT NULL,
    liked_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    PRIMARY KEY (post_owner, post_uuid, liked_by),
    FOREIGN KEY (post_owner, post_uuid) REFERENCES post(owner, uuid)
);

CREATE MATERIALIZED VIEW stats AS
    WITH pre AS (
        SELECT post_owner as owner, post_uuid as uuid, liked_at::DATE as date, SUM(weight) as updv
        from "like"
        group by owner, uuid, date
    ),
         r AS (
             SELECT owner, uuid, date, SUM(updv) OVER (PARTITION BY (owner, uuid) ORDER BY date) as updv
             from pre
         )
    SELECT owner, uuid, json_object_agg(date, updv) AS stats FROM r
    GROUP BY owner, uuid;

CREATE UNIQUE INDEX stats_pk_idx ON stats(owner, uuid);

CREATE MATERIALIZED VIEW pdv_stats AS
    WITH pre AS (
        SELECT address, timestamp::DATE as date, SUM(updv) as updv
        from updv
        group by address, date
    ),
     r AS (
         SELECT address, date, SUM(updv) OVER (PARTITION BY (address) ORDER BY date) as updv
         from pre
     )
SELECT address, json_object_agg(date, updv) AS stats FROM r
GROUP BY address;

CREATE UNIQUE INDEX pdv_stats_pk_idx ON pdv_stats(address);

CREATE MATERIALIZED VIEW calculated_post AS
    SELECT owner, uuid, title, category, preview_image, text, post.created_at,
        COALESCE(COUNT(weight) FILTER (WHERE weight = 1), 0) as likes,
        COALESCE(COUNT(weight) FILTER (WHERE weight = -1), 0) AS dislikes,
        COALESCE(SUM(weight), 0) AS updv
    FROM post
    LEFT JOIN "like" ON post.owner = "like".post_owner AND post.uuid = "like".post_uuid
    WHERE deleted_at IS NULL
    GROUP BY owner, uuid, title, category, preview_image, text, post.created_at;

CREATE UNIQUE INDEX calculated_post_pk_idx ON calculated_post(owner, uuid);
CREATE INDEX calculated_post_created_at_idx ON calculated_post(created_at DESC);
CREATE INDEX calculated_post_likes_idx ON calculated_post(likes DESC);
CREATE INDEX calculated_post_category_idx ON calculated_post(category);

CREATE OR REPLACE FUNCTION set_profile() RETURNS TRIGGER AS $set_profile$
DECLARE
    query JSONB := to_jsonb(NEW);
BEGIN
    IF (TG_OP = 'INSERT') THEN
        IF query ? 'owner' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.owner, '0001-01-01'::DATE) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'post_owner' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.post_owner, '0001-01-01'::DATE) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'liked_by' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.liked_by, '0001-01-01'::DATE) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'follower' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.follower, '0001-01-01'::DATE) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'followee' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.followee, '0001-01-01'::DATE) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'address' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.address, '0001-01-01'::DATE) ON CONFLICT(address) DO NOTHING;
        END IF;

        RETURN NEW;
    ELSEIF (TG_OP = 'UPDATE') THEN
        IF query ? 'deleted_by' THEN
            IF NEW.deleted_by IS NOT NULL THEN
                INSERT INTO profile(address, created_at) VALUES(NEW.deleted_by, '0001-01-01'::DATE) ON CONFLICT(address) DO NOTHING;
            END IF;
        END IF;

        RETURN NEW;
    END IF;

    RETURN NULL;
END;
$set_profile$ LANGUAGE plpgsql;

CREATE TRIGGER add_missed_profile_post_trigger BEFORE INSERT OR UPDATE ON post FOR EACH ROW EXECUTE PROCEDURE set_profile();
CREATE TRIGGER add_missed_profile_like_trigger BEFORE INSERT ON "like" FOR EACH ROW EXECUTE PROCEDURE set_profile();
CREATE TRIGGER add_missed_profile_follow_trigger BEFORE INSERT ON follow FOR EACH ROW EXECUTE PROCEDURE set_profile();
CREATE TRIGGER add_missed_profile_updv_trigger BEFORE INSERT ON updv FOR EACH ROW EXECUTE PROCEDURE set_profile();

COMMIT;