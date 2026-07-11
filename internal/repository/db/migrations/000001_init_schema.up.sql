CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_identity_id" varchar,
  "email" varchar,
  "password_hash" varchar,
  "role" varchar(5) NOT NULL,
  "email_validated" bool NOT NULL DEFAULT false,
  "is_active" bool NOT NULL DEFAULT true,
  "secret_token_key" varchar(64) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT Now(),
  "updated_at" timestamp
);

-- Enforce unique emails case-insensitively, but only for users that have one.
-- Anonymous users keep email NULL and are exempt (a partial index allows many NULLs).
CREATE UNIQUE INDEX users_email_key ON "users" (lower("email")) WHERE "email" IS NOT NULL;

CREATE TABLE "time_entries" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_id" uuid NOT NULL,
  "tag" varchar NOT NULL,
  "time_start" timestamp NOT NULL,
  "time_end" timestamp,
  "created_at" timestamp NOT NULL DEFAULT Now(),
  "updated_at" timestamp
);

ALTER TABLE "time_entries" ADD FOREIGN KEY ("user_id") REFERENCES "users" ("id");
