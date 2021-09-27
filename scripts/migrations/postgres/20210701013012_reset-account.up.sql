BEGIN;

ALTER TABLE "like" DROP CONSTRAINT IF EXISTS like_post_owner_post_uuid_fkey;
ALTER TABLE "like" DROP CONSTRAINT IF EXISTS like_post_owner_fkey1;
ALTER TABLE "like" ADD FOREIGN KEY (post_owner, post_uuid) REFERENCES post(owner, uuid) ON UPDATE CASCADE;

COMMIT;