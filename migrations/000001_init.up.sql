CREATE OR REPLACE FUNCTION update_updated_at()
    RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;  
END;
$$ LANGUAGE plpgsql;

CREATE TABLE IF NOT EXISTS public.users
(
  id bigserial NOT NULL PRIMARY KEY,

  uuid text DEFAULT REPLACE(gen_random_uuid()::text, '-', '' ),
  username text NOT NULL,

  api_key text NOT NULL,

  created_at timestamptz DEFAULT NOW() NOT NULL,
  updated_at timestamptz DEFAULT NOW() NOT NULL
);

CREATE TRIGGER users_updated_at_update_trigger
  BEFORE UPDATE
  ON public.users
  FOR EACH ROW  
EXECUTE PROCEDURE update_updated_at();

CREATE TYPE deployment_status AS ENUM ('PENDING', 'READY', 'DELETING');

CREATE TABLE IF NOT EXISTS public.deployments (
  id bigserial NOT NULL PRIMARY KEY,
  uuid text NOT NULL DEFAULT replace(gen_random_uuid ()::text, '-', ''),
  user_id bigserial NOT NULL CONSTRAINT deployments_user_id_fkey REFERENCES public.users (id) ON UPDATE CASCADE ON DELETE RESTRICT,
  
  sub_domain TEXT UNIQUE NOT NULL,
  image_tag TEXT NOT NULL,
  container_id TEXT UNIQUE DEFAULT NULL,

  port INTEGER UNIQUE DEFAULT NULL,

  status deployment_status DEFAULT 'PENDING',
  
  created_at timestamptz DEFAULT now() NOT NULL,
  updated_at timestamptz DEFAULT now() NOT NULL
);

CREATE TRIGGER deployments_updated_at_update_trigger
  BEFORE UPDATE
  ON public.deployments
  FOR EACH ROW  
EXECUTE PROCEDURE update_updated_at();
