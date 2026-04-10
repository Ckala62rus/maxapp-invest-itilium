-- Stores linked MAX users and registration state.
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    username TEXT NOT NULL DEFAULT '',
    full_name TEXT NOT NULL,
    department TEXT NOT NULL,
    employee_found BOOLEAN NOT NULL DEFAULT FALSE,
    registration_required BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
