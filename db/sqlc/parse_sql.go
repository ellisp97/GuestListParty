package db

import (
	"context"
	"database/sql"
)

// getGuestFromSQLQuery returns a Guest object following a CreateGuest action
// by utilising the LastInsertID field, this is because MySQL has no concept
// of `RETURNING` which sqlc can implement
func (q *Queries) getGuestFromSQLQuery(query sql.Result) (Guest, error) {
	var guest Guest

	id, err := query.LastInsertId()
	if err != nil {
		return guest, err
	}
	guest, err = q.GetGuest(context.Background(), int32(id))
	return guest, err
}

// getTableFromSQLQuery returns a Table object following a CreateTable action
func (q *Queries) getTableFromSQLQuery(query sql.Result) (Table, error) {
	var table Table

	id, err := query.LastInsertId()
	if err != nil {
		return table, err
	}
	table, err = q.GetTable(context.Background(), int32(id))
	return table, err
}

// getArrivalFromSQLQuery returns a Arrival object following a CreateArrival action
func (q *Queries) getArrivalFromSQLQuery(query sql.Result) (Arrival, error) {
	var arrival Arrival

	id, err := query.LastInsertId()
	if err != nil {
		return arrival, err
	}
	arrival, err = q.GetArrival(context.Background(), int32(id))
	return arrival, err
}
