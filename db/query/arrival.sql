-- name: CreateArrival :execresult
INSERT INTO arrivals (
    guest_id,
    table_id,
    party_size
) VALUES (
    ?, ?, ?
);

-- name: GetArrival :one
SELECT * from arrivals
WHERE id =? LIMIT 1;

-- name: GetArrivalFromGuest :one
SELECT * from arrivals
WHERE guest_id =?;

-- name: GetArrivals :many
SELECT * FROM arrivals
WHERE   
    guest_id = ? OR
    table_id = ?
ORDER BY id
LIMIT ?
OFFSET ?;