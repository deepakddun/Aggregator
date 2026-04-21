CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE feed_follows (
     id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
     created_at timestamptz NOT NULL,
     updated_at timestamptz not null DEFAULT now(),
     user_id uuid not null , 
     feed_id uuid not null,
     CONSTRAINT fk1 FOREIGN KEY(user_id)
     REFERENCES users(id)
     ON DELETE CASCADE,
     CONSTRAINT fk2 FOREIGN KEY(feed_id)
     REFERENCES feeds(id)
     ON DELETE CASCADE,
     CONSTRAINT fk3 UNIQUE(feed_id , user_id)
     
);


