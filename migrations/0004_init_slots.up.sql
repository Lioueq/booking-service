CREATE TABLE IF NOT EXISTS slots (
    slot_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    room_id UUID NOT NULL REFERENCES rooms(room_id) ON DELETE CASCADE,
    start_at TIMESTAMPTZ NOT NULL,
    end_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (end_at > start_at),
    UNIQUE (room_id, start_at, end_at)
);

CREATE INDEX IF NOT EXISTS idx_slots_room_start_at ON slots (room_id, start_at);
