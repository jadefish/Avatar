BEGIN;

CREATE TABLE "public"."accounts" (
    "id" serial PRIMARY KEY ,
    "created_at" timestamptz NOT NULL DEFAULT NOW(),
    "updated_at" timestamptz NOT NULL DEFAULT NOW(),
    "deleted_at" timestamptz DEFAULT NULL COMMENT,
    "last_logged_in_at" timestamptz DEFAULT NULL COMMENT 'UTC timestamp of last successful login',
    "name" varchar(20) NOT NULL DEFAULT NULL UNIQUE COMMENT 'Account name',
    "email" validemail NOT NULL DEFAULT NULL UNIQUE COMMENT 'Account owner''s registered email address',
    "password_hash" text NOT NULL DEFAULT NULL COMMENT 'Password hash',
    "creation_ip" inet NOT NULL DEFAULT NULL COMMENT 'IP address of record creation',
    "last_login_ip" inet DEFAULT NULL COMMENT 'IP address of last successful login'
);

CREATE TRIGGER "accounts_update_updated_at"
    BEFORE UPDATE ON "accounts"
    FOR EACH ROW EXECUTE PROCEDURE
        on_update_set_updated_at();

CREATE INDEX accounts_email_index ON "accounts"("email");
CREATE INDEX accounts_name_index ON "accounts"("name");

COMMIT;
