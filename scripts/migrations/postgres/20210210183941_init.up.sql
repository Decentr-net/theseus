BEGIN;

CREATE TABLE height (
    height BIGINT
);

INSERT INTO height VALUES(0);

CREATE TABLE profile (
    address TEXT PRIMARY KEY,
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    avatar TEXT NOT NULL DEFAULT '',
    gender TEXT NOT NULL DEFAULT '',
    birthday TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMP WITHOUT TIME ZONE NOT NULL
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

CREATE INDEX post_created_at_idx ON post(created_at DESC);
CREATE INDEX post_category_idx ON post(category);

CREATE TABLE "like" (
    post_owner TEXT NOT NULL REFERENCES profile(address),
    post_uuid TEXT NOT NULL,
    liked_by TEXT NOT NULL REFERENCES profile(address),
    weight INT2 NOT NULL,
    liked_at TIMESTAMP WITHOUT TIME ZONE NOT NULL,

    PRIMARY KEY (post_owner, post_uuid, liked_by),
    FOREIGN KEY (post_owner, post_uuid) REFERENCES post(owner, uuid)
);

CREATE OR REPLACE FUNCTION set_profile() RETURNS TRIGGER AS $set_profile$
DECLARE
    query JSONB := to_jsonb(NEW);
BEGIN
    IF (TG_OP = 'INSERT') THEN
        IF query ? 'owner' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.owner, NOW()) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'post_owner' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.post_owner, NOW()) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'liked_by' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.liked_by, NOW()) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'follower' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.follower, NOW()) ON CONFLICT(address) DO NOTHING;
        END IF;

        IF query ? 'followee' THEN
            INSERT INTO profile(address, created_at) VALUES(NEW.followee, NOW()) ON CONFLICT(address) DO NOTHING;
        END IF;

        RETURN NEW;
    ELSEIF (TG_OP = 'UPDATE') THEN
        IF query ? 'deleted_by' THEN
            IF NEW.deleted_by IS NOT NULL THEN
                INSERT INTO profile(address, created_at) VALUES(NEW.deleted_by, NOW()) ON CONFLICT(address) DO NOTHING;
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

COMMIT;