BEGIN;

ALTER TABLE "like" DROP CONSTRAINT like_post_owner_post_uuid_fkey;
ALTER TABLE "like" ADD FOREIGN KEY (post_owner, post_uuid) REFERENCES post(owner, uuid);

COMMIT;