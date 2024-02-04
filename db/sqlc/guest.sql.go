// Code generated by sqlc. DO NOT EDIT.
// source: guest.sql

package db

import (
	"context"
	"database/sql"
)

const createGuest = `-- name: CreateGuest :execresult
INSERT INTO guests(
    guest_name,
    entourage,
    table_id,
    arrival_time
) VALUES (
    ?, ?, ?, ?
)
`

type CreateGuestParams struct {
	GuestName   string       `json:"guest_name"`
	Entourage   int32        `json:"entourage"`
	TableID     int32        `json:"table_id"`
	ArrivalTime sql.NullTime `json:"arrival_time"`
}

func (q *Queries) CreateGuest(ctx context.Context, arg CreateGuestParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, createGuest,
		arg.GuestName,
		arg.Entourage,
		arg.TableID,
		arg.ArrivalTime,
	)
}

const deleteGuest = `-- name: DeleteGuest :exec
DELETE FROM guests
WHERE id =?
`

func (q *Queries) DeleteGuest(ctx context.Context, id int32) error {
	_, err := q.db.ExecContext(ctx, deleteGuest, id)
	return err
}

const getArrivedGuests = `-- name: GetArrivedGuests :many
SELECT id, guest_name, entourage, table_id, arrival_time, created_at FROM guests
WHERE id IN (
    SELECT guest_id FROM arrivals
)
ORDER BY arrival_time
LIMIT ?
OFFSET ?
`

type GetArrivedGuestsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetArrivedGuests(ctx context.Context, arg GetArrivedGuestsParams) ([]Guest, error) {
	rows, err := q.db.QueryContext(ctx, getArrivedGuests, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Guest{}
	for rows.Next() {
		var i Guest
		if err := rows.Scan(
			&i.ID,
			&i.GuestName,
			&i.Entourage,
			&i.TableID,
			&i.ArrivalTime,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const getEmptySeats = `-- name: GetEmptySeats :one
SELECT (SELECT IFNULL(SUM(size), 0) from tables) - (SELECT IFNULL(SUM(occupied), 0) from tables) AS seats_empty
`

func (q *Queries) GetEmptySeats(ctx context.Context) (int32, error) {
	row := q.db.QueryRowContext(ctx, getEmptySeats)
	var seats_empty int32
	err := row.Scan(&seats_empty)
	return seats_empty, err
}

const getGuest = `-- name: GetGuest :one
SELECT id, guest_name, entourage, table_id, arrival_time, created_at FROM guests
WHERE id = ? LIMIT 1
`

func (q *Queries) GetGuest(ctx context.Context, id int32) (Guest, error) {
	row := q.db.QueryRowContext(ctx, getGuest, id)
	var i Guest
	err := row.Scan(
		&i.ID,
		&i.GuestName,
		&i.Entourage,
		&i.TableID,
		&i.ArrivalTime,
		&i.CreatedAt,
	)
	return i, err
}

const getGuestForUpdate = `-- name: GetGuestForUpdate :one
SELECT id, guest_name, entourage, table_id, arrival_time, created_at FROM guests
WHERE id = ? LIMIT 1
FOR UPDATE
`

func (q *Queries) GetGuestForUpdate(ctx context.Context, id int32) (Guest, error) {
	row := q.db.QueryRowContext(ctx, getGuestForUpdate, id)
	var i Guest
	err := row.Scan(
		&i.ID,
		&i.GuestName,
		&i.Entourage,
		&i.TableID,
		&i.ArrivalTime,
		&i.CreatedAt,
	)
	return i, err
}

const getGuestFromName = `-- name: GetGuestFromName :one
SELECT id, guest_name, entourage, table_id, arrival_time, created_at FROM guests
WHERE guest_name = ? LIMIT 1
`

func (q *Queries) GetGuestFromName(ctx context.Context, guestName string) (Guest, error) {
	row := q.db.QueryRowContext(ctx, getGuestFromName, guestName)
	var i Guest
	err := row.Scan(
		&i.ID,
		&i.GuestName,
		&i.Entourage,
		&i.TableID,
		&i.ArrivalTime,
		&i.CreatedAt,
	)
	return i, err
}

const getGuests = `-- name: GetGuests :many
SELECT id, guest_name, entourage, table_id, arrival_time, created_at FROM guests 
ORDER BY id
LIMIT ?
OFFSET ?
`

type GetGuestsParams struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

func (q *Queries) GetGuests(ctx context.Context, arg GetGuestsParams) ([]Guest, error) {
	rows, err := q.db.QueryContext(ctx, getGuests, arg.Limit, arg.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Guest{}
	for rows.Next() {
		var i Guest
		if err := rows.Scan(
			&i.ID,
			&i.GuestName,
			&i.Entourage,
			&i.TableID,
			&i.ArrivalTime,
			&i.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, i)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

const updateGuestArrival = `-- name: UpdateGuestArrival :exec
UPDATE guests
SET entourage = ?
WHERE id = ?
`

type UpdateGuestArrivalParams struct {
	Entourage int32 `json:"entourage"`
	ID        int32 `json:"id"`
}

func (q *Queries) UpdateGuestArrival(ctx context.Context, arg UpdateGuestArrivalParams) error {
	_, err := q.db.ExecContext(ctx, updateGuestArrival, arg.Entourage, arg.ID)
	return err
}
