CREATE OR REPLACE FUNCTION update_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;  
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS "users"
(
  "id" serial NOT NULL PRIMARY KEY,

  "uuid" text DEFAULT gen_random_uuid(),
  "name" text NOT NULL,

  "api_key" text NOT NULL,

  "created_at" timestamptz DEFAULT NOW() NOT NULL,
  "updated_at" timestamptz DEFAULT NOW() NOT NULL  
);

CREATE TRIGGER users_updated_at_update_trigger
  BEFORE UPDATE
  ON "users"
  FOR EACH ROW  
EXECUTE PROCEDURE update_updated_at();
