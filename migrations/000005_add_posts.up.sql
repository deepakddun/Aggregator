CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE posts (
     id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
     url VARCHAR(100) UNIQUE not null ,
     title VARCHAR(100),
     description text,
     created_at timestamptz NOT NULL,
     updated_at timestamptz not null DEFAULT now(),
     published_at timestamptz,
     feed_id uuid,
     CONSTRAINT fk FOREIGN KEY(feed_id)
     REFERENCES feeds(id)
     ON DELETE CASCADE
);





