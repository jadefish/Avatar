BEGIN;

CREATE TABLE "public"."shards" (
    "id" serial PRIMARY KEY,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW(),
    "deleted_at" timestamptz DEFAULT NULL,
    "name" varchar(32) NOT NULL,
    "time_zone" timezone NOT NULL DEFAULT 'Etc/UTC',
    "capacity" int2 NOT NULL DEFAULT 3000 CHECK("capacity" > 0),
    "ip_address" inet NOT NULL
);

COMMENT ON COLUMN "public"."shards"."name"
    IS 'Name of the shard';
COMMENT ON COLUMN "public"."shards"."time_zone"
    IS 'Name of the shard''s time zone (see tzdata)';
COMMENT ON COLUMN "public"."shards"."capacity"
    IS 'Maximum number of players allowed to concurrently exist in the game world';
COMMENT ON COLUMN "public"."shards"."ip_address"
    IS 'IP address of the shard''s game server';

CREATE TRIGGER "shards_update_updated_at"
    BEFORE UPDATE ON "shards"
    FOR EACH ROW EXECUTE PROCEDURE
        on_update_set_updated_at();

COMMIT;
