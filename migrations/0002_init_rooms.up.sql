CREATE TABLE IF NOT EXISTS rooms(
    room_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    room_name VARCHAR(255) UNIQUE NOT NULL,
    room_description VARCHAR(255),
    capacity INTEGER CHECK (capacity >= 0),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
