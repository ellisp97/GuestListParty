// Code generated by sqlc. DO NOT EDIT.
// source: arrival.sql

package db

import (
	"context"
	"database/sql"
)

const createArrival = `-- name: CreateArrival :execresult
INSERT INTO arrivals (
    guest_id,
    table_id,
    party_size
) VALUES (
    ?, ?, ?
)
`

type CreateArrivalParams struct {
	GuestID   int32 `json:"guest_id"`
	TableID   int32 `json:"table_id"`
	PartySize int32 `json:"party_size"`
}

func (q *Queries) CreateArrival(ctx context.Context, arg CreateArrivalParams) (sql.Result, error) {
	return q.db.ExecContext(ctx, createArrival, arg.GuestID, arg.TableID, arg.PartySize)
}

const getArrival = `-- name: GetArrival :one
SELECT id, guest_id, table_id, party_size from arrivals
WHERE id =? LIMIT 1
`

func (q *Queries) GetArrival(ctx context.Context, id int32) (Arrival, error) {
	row := q.db.QueryRowContext(ctx, getArrival, id)
	var i Arrival
	err := row.Scan(
		&i.ID,
		&i.GuestID,
		&i.TableID,
		&i.PartySize,
	)
	return i, err
}

const getArrivalFromGuest = `-- name: GetArrivalFromGuest :one
SELECT id, guest_id, table_id, party_size from arrivals
WHERE guest_id =?
`

func (q *Queries) GetArrivalFromGuest(ctx context.Context, guestID int32) (Arrival, error) {
	row := q.db.QueryRowContext(ctx, getArrivalFromGuest, guestID)
	var i Arrival
	err := row.Scan(
		&i.ID,
		&i.GuestID,
		&i.TableID,
		&i.PartySize,
	)
	return i, err
}

const getArrivals = `-- name: GetArrivals :many
SELECT id, guest_id, table_id, party_size FROM arrivals
WHERE   
    guest_id = ? OR
    table_id = ?
ORDER BY id
LIMIT ?
OFFSET ?
`

type GetArrivalsParams struct {
	GuestID int32 `json:"guest_id"`
	TableID int32 `json:"table_id"`
	Limit   int32 `json:"limit"`
	Offset  int32 `json:"offset"`
}

func (q *Queries) GetArrivals(ctx context.Context, arg GetArrivalsParams) ([]Arrival, error) {
	rows, err := q.db.QueryContext(ctx, getArrivals,
		arg.GuestID,
		arg.TableID,
		arg.Limit,
		arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Arrival{}
	for rows.Next() {
		var i Arrival
		if err := rows.Scan(
			&i.ID,
			&i.GuestID,
			&i.TableID,
			&i.PartySize,
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
