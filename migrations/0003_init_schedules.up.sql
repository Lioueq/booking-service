CREATE TABLE IF NOT EXISTS schedules (
    schedule_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    room_id UUID NOT NULL UNIQUE REFERENCES rooms(room_id) ON DELETE CASCADE,
    days_of_week INT[] NOT NULL,
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_time > start_time)
);