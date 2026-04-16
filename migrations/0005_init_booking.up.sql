CREATE TABLE IF NOT EXISTS bookings (
    booking_id UUID PRIMARY KEY DEFAULT gen_random_uuid() NOT NULL,
    slot_id UUID NOT NULL REFERENCES slots(slot_id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(user_id),
    status VARCHAR(255) NOT NULL CHECK (status IN ('active', 'cancelled')),
    conference_link VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_bookings_status
ON bookings (slot_id)
WHERE status = 'active';
