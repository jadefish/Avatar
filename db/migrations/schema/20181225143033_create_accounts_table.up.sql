BEGIN;

CREATE TABLE "public"."accounts" (
    "id" serial PRIMARY KEY,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW(),
    "deleted_at" timestamptz DEFAULT NULL,
    "last_logged_in_at" timestamptz DEFAULT NULL,
    "name" varchar(20) NOT NULL DEFAULT NULL UNIQUE,
    "email" validemail NOT NULL DEFAULT NULL UNIQUE,
    "password_hash" text NOT NULL DEFAULT NULL,
    "creation_ip" inet NOT NULL DEFAULT NULL,
    "last_login_ip" inet DEFAULT NULL
);

COMMENT ON COLUMN "public"."accounts"."last_logged_in_at"
    IS 'UTC timestamp of last successful login';
COMMENT ON COLUMN "public"."accounts"."name"
    IS  'Account name';
COMMENT ON COLUMN "public"."accounts"."email"
    IS 'Account owner''s registered email address';
COMMENT ON COLUMN "public"."accounts"."password_hash"
    IS 'Password hash';
COMMENT ON COLUMN "public"."accounts"."creation_ip"
    IS 'IP address of record creation';
COMMENT ON COLUMN "public"."accounts"."last_login_ip"
    IS 'IP address of last successful login';

CREATE TRIGGER "accounts_update_updated_at"
    BEFORE UPDATE ON "accounts"
    FOR EACH ROW EXECUTE PROCEDURE
        on_update_set_updated_at();

CREATE INDEX accounts_email_index ON "accounts"("email");
CREATE INDEX accounts_name_index ON "accounts"("name");

COMMIT;
