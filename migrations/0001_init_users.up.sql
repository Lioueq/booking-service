CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users(
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    user_email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    user_role VARCHAR(255) NOT NULL CHECK (user_role IN ('admin', 'user')),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

INSERT INTO users (user_id, user_email, password_hash, user_role, created_at) VALUES
('11111111-1111-1111-1111-111111111111', 'testuser@test.test', 'test1', 'user', NOW()),
('11111111-1111-1111-1111-111111111112', 'testadmin@test.test', 'test2', 'admin', NOW());
