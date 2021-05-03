BEGIN;

ALTER TABLE updv DROP CONSTRAINT updv_address_fkey;
ALTER TABLE follow DROP CONSTRAINT follow_followee_fkey;
ALTER TABLE follow DROP CONSTRAINT follow_follower_fkey;
ALTER TABLE post DROP CONSTRAINT post_owner_fkey;
ALTER TABLE post DROP CONSTRAINT post_deleted_by_fkey;
ALTER TABLE "like" DROP CONSTRAINT like_liked_by_fkey;
ALTER TABLE "like" DROP CONSTRAINT like_post_owner_fkey;

DROP FUNCTION set_profile() CASCADE;

DROP TABLE profile;

COMMIT;