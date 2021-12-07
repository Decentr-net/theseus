BEGIN;

DROP TRIGGER trigger_post_unique_slug ON post;

ALTER TABLE post
    DROP COLUMN slug;

DROP FUNCTION IF EXISTS unique_post_slug();

DROP MATERIALIZED VIEW calculated_post;
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

COMMIT;