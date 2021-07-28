BEGIN;

ALTER TABLE "like" DROP CONSTRAINT like_post_owner_fkey1;
ALTER TABLE "like" ADD FOREIGN KEY (post_owner, post_uuid) REFERENCES post(owner, uuid) ON UPDATE CASCADE;

COMMIT;