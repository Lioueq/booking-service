BEGIN;


INSERT INTO users (user_email, password_hash, user_role)
SELECT
  format('load_user_%s@test.local', gs),
  'load-hash',
  'user'
FROM generate_series(1, 10000) AS gs
ON CONFLICT (user_email) DO NOTHING;

TRUNCATE TABLE bookings, slots, schedules, rooms RESTART IDENTITY CASCADE;

-- Up to 50 rooms.
INSERT INTO rooms (room_name, room_description, capacity)
SELECT
  format('Load Room %s', gs),
  format('Load room #%s', gs),
  4 + (gs % 16)
FROM generate_series(1, 50) AS gs;

INSERT INTO schedules (room_id, days_of_week, start_time, end_time)
SELECT room_id, ARRAY[1,2,3,4,5,6,7], '09:00', '19:00'
FROM rooms;

-- 100 days * 50 rooms * 20 slots/day = 100,000 slots.
WITH target_rooms AS (
  SELECT room_id
  FROM rooms
  ORDER BY room_name
  LIMIT 50
),
days AS (
  SELECT generate_series((CURRENT_DATE + INTERVAL '1 day')::date, (CURRENT_DATE + INTERVAL '100 day')::date, INTERVAL '1 day')::date AS d
),
slot_idx AS (
  SELECT generate_series(0, 19) AS i
)
INSERT INTO slots (room_id, start_at, end_at)
SELECT
  r.room_id,
  make_timestamptz(
    EXTRACT(YEAR FROM d.d)::int,
    EXTRACT(MONTH FROM d.d)::int,
    EXTRACT(DAY FROM d.d)::int,
    9,
    0,
    0,
    'UTC'
  ) + (si.i * INTERVAL '30 minutes') AS start_at,
  make_timestamptz(
    EXTRACT(YEAR FROM d.d)::int,
    EXTRACT(MONTH FROM d.d)::int,
    EXTRACT(DAY FROM d.d)::int,
    9,
    0,
    0,
    'UTC'
  ) + ((si.i + 1) * INTERVAL '30 minutes') AS end_at
FROM target_rooms r
CROSS JOIN days d
CROSS JOIN slot_idx si
ON CONFLICT (room_id, start_at, end_at) DO NOTHING;

-- Up to 100k bookings, one booking per slot.
WITH users_pool AS (
  SELECT user_id, row_number() OVER (ORDER BY user_id) AS rn
  FROM users
  WHERE user_role = 'user'
),
users_count AS (
  SELECT COUNT(*) AS cnt FROM users_pool
),
slots_seq AS (
  SELECT slot_id, row_number() OVER (ORDER BY start_at, slot_id) AS rn
  FROM slots
  ORDER BY start_at, slot_id
  LIMIT 100000
)
INSERT INTO bookings (slot_id, user_id, status, conference_link)
SELECT
  s.slot_id,
  u.user_id,
  CASE WHEN s.rn % 10 < 7 THEN 'active' ELSE 'cancelled' END,
  CASE WHEN s.rn % 9 = 0 THEN format('https://conf.load/%s', s.slot_id) ELSE NULL END
FROM slots_seq s
CROSS JOIN users_count uc
JOIN users_pool u ON u.rn = ((s.rn - 1) % uc.cnt) + 1;

COMMIT;
