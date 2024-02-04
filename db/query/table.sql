-- name: CreateTable :execresult
INSERT INTO tables(
    size,
    occupied
) VALUES (
    ?, ?
);

-- name: GetTables :many
SELECT * FROM tables 
ORDER BY id
LIMIT ?
OFFSET ?;

-- name: GetTable :one
SELECT * FROM tables
WHERE id = ? LIMIT 1; 

-- name: GetTableForUpdate :one
SELECT * FROM tables
WHERE id = ? LIMIT 1
FOR UPDATE;

-- name: UpdateTable :exec
UPDATE tables
SET size = ?,
occupied = ?
WHERE id = ?;

-- name: DeleteTable :exec
DELETE FROM tables
WHERE id=?;