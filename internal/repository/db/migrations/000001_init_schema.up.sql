CREATE TABLE "users" (
  "id" uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  "user_identity_id" uuid,
  "email" varchar,
  "role" varchar(5) NOT NULL,
  "email_validated" bool NOT NULL DEFAULT false,
  "is_active" bool NOT NULL DEFAULT true,
  "secret_token_key" varchar(64) NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT Now(),
  "updated_at" timestamp
);

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
