CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE feeds (
     id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
     name VARCHAR(100) not null ,
     url VARCHAR(100) UNIQUE not null ,
     created_at timestamptz NOT NULL,
     updated_at timestamptz not null DEFAULT now(),
     user_id uuid not null , 
     CONSTRAINT fk FOREIGN KEY(user_id)
     REFERENCES users(id)
     ON DELETE CASCADE
);

CREATE INDEX idx_feeds_name ON feeds (name);
