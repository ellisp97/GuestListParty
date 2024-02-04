-- name: CreateGuest :execresult
INSERT INTO guests(
    guest_name,
    entourage,
    table_id,
    arrival_time
) VALUES (
    ?, ?, ?, ?
);

-- name: GetGuests :many
SELECT * FROM guests 
ORDER BY id
LIMIT ?
OFFSET ?;

-- name: GetGuest :one
SELECT * FROM guests
WHERE id = ? LIMIT 1;

-- name: GetGuestForUpdate :one
SELECT * FROM guests
WHERE id = ? LIMIT 1
FOR UPDATE;

-- name: UpdateGuestArrival :exec
UPDATE guests
SET entourage = ?
WHERE id = ?;

-- name: DeleteGuest :exec
DELETE FROM guests
WHERE id =?;

-- name: GetEmptySeats :one
SELECT (SELECT IFNULL(SUM(size), 0) from tables) - (SELECT IFNULL(SUM(occupied), 0) from tables) AS seats_empty;

-- name: GetGuestFromName :one
SELECT * FROM guests
WHERE guest_name = ? LIMIT 1;

-- name: GetArrivedGuests :many
SELECT * FROM guests
WHERE id IN (
    SELECT guest_id FROM arrivals
)
ORDER BY arrival_time
LIMIT ?
OFFSET ?;