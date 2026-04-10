CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
     id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
     name VARCHAR(100) not null ,
     created_at timestamptz NOT NULL,
     updated_at timestamptz not null DEFAULT now()
);

CREATE INDEX idx_user_name ON users (name);
