BEGIN;

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

ALTER TABLE post
    ADD COLUMN slug TEXT NULL UNIQUE;

CREATE OR REPLACE FUNCTION unique_post_slug()
    RETURNS TRIGGER AS
$$

DECLARE
    key   TEXT;
    qry   TEXT;
    found TEXT;
BEGIN

    -- generate the first part of a query as a string with safely
    -- escaped table name, using || to concat the parts
    qry := 'SELECT slug FROM ' || quote_ident(TG_TABLE_NAME) || ' WHERE slug=';

    -- This loop will probably only run once per call until we've generated
    -- millions of ids.
    LOOP

        -- Generate our string bytes and re-encode as a base64 string.
        key := encode(gen_random_bytes(6), 'base64');

        -- Base64 encoding contains 2 URL unsafe characters by default.
        -- The URL-safe version has these replacements.
        key := replace(key, '/', '_'); -- url safe replacement
        key := replace(key, '+', '-');
        -- url safe replacement

        -- Concat the generated key (safely quoted) with the generated query
        -- and run it.
        -- SELECT id FROM "test" WHERE id='blahblah' INTO found
        -- Now "found" will be the duplicated id or NULL.
        EXECUTE qry || quote_literal(key) INTO found;

        -- Check to see if found is NULL.
        -- If we checked to see if found = NULL it would always be FALSE
        -- because (NULL = NULL) is always FALSE.
        IF found IS NULL THEN

            -- If we didn't find a collision then leave the LOOP.
            EXIT;
        END IF;

        -- We haven't EXITed yet, so return to the top of the LOOP
        -- and try again.
    END LOOP;

    NEW.slug = key;

    -- The RECORD returned here is what will actually be INSERTed,
    -- or what the next trigger will get if there is one.
    RETURN NEW;
END;
$$ LANGUAGE 'plpgsql';

CREATE TRIGGER trigger_post_unique_slug
    BEFORE UPDATE
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE unique_post_slug();

-- Dummy update to invoke the trigger_post_unique_slug
UPDATE post
SET uuid = uuid;

-- Existing request updated, drop the trigger
DROP TRIGGER trigger_post_unique_slug ON post;

-- Recreate the trigger but apply it only on INSERT
CREATE TRIGGER trigger_post_unique_slug
    BEFORE INSERT
    ON post
    FOR EACH ROW
EXECUTE PROCEDURE unique_post_slug();

ALTER TABLE post
    ALTER COLUMN slug SET NOT NULL;

-- Recreate DROP MATERIALIZED VIEW calculated_post with adding slug
DROP MATERIALIZED VIEW calculated_post;

CREATE MATERIALIZED VIEW calculated_post AS
SELECT owner, uuid, title, category, preview_image, text, post.created_at,
       COALESCE(COUNT(weight) FILTER (WHERE weight = 1), 0) as likes,
       COALESCE(COUNT(weight) FILTER (WHERE weight = -1), 0) AS dislikes,
       COALESCE(SUM(weight), 0) AS updv,
       slug
FROM post
         LEFT JOIN "like" ON post.owner = "like".post_owner AND post.uuid = "like".post_uuid
WHERE deleted_at IS NULL
GROUP BY owner, uuid, title, category, preview_image, text, post.created_at;

CREATE UNIQUE INDEX calculated_post_pk_idx ON calculated_post(owner, uuid);
CREATE INDEX calculated_post_created_at_idx ON calculated_post(created_at DESC);
CREATE INDEX calculated_post_likes_idx ON calculated_post(likes DESC);
CREATE INDEX calculated_post_category_idx ON calculated_post(category);
CREATE INDEX calculated_post_slug_idx ON calculated_post(slug);

COMMIT;