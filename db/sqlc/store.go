package db

import (
	"context"
	"database/sql"
	"fmt"
)

type Store interface {
	Querier
	AssignTableTx(ctx context.Context, arg AssignTableTxParams) (AssignTableTxResult, error)
	DeleteGuestTx(ctx context.Context, id int32) error
}

// Store provides all functions to execute db queries and transactions
type SQLStore struct {
	*Queries
	db *sql.DB
}

func NewStore(db *sql.DB) *SQLStore {
	return &SQLStore{
		db:      db,
		Queries: New(db),
	}
}

// execTx executes function within a db transaction
func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
		}
		return err
	}
	return tx.Commit()
}

// AssignTableParams contains input parameters of the transaction assigning a guest to a table
type AssignTableTxParams struct {
	UserID       int64 `json:"user_id"`
	NewEntourage int64 `json:"new_entourage"`
	TableID      int64 `json:"table_id"`
}

// AssignTableTxResult contains result of the assign table transaction
type AssignTableTxResult struct {
	Arrival  Arrival `json:"arrival"`
	Table    Table   `json:"table"`
	Guest    Guest   `json:"guest"`
	OldTable Table   `json:"old_table"`
}

var txKey = struct{}{}

func InsufficientTableSizeErr(tableID int) error {
	return fmt.Errorf("Table %d has insufficient space", tableID)
}

// AssignTableTx assigns a guest alongwith their entourage to a table, and will return an error if the party size is bigger than the table size
func (store *SQLStore) AssignTableTx(ctx context.Context, arg AssignTableTxParams) (AssignTableTxResult, error) {
	var result AssignTableTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		result.Guest, err = q.GetGuestForUpdate(ctx, int32(arg.UserID))
		if err != nil {
			return err
		}

		// Must also check that the guest is not already arrived by accessing the arrivals table
		_, err = q.GetArrivalFromGuest(ctx, int32(arg.UserID))
		if err == nil {
			return fmt.Errorf("An arrival has already been made for this guest")
		}

		result.Table, err = q.GetTableForUpdate(ctx, int32(arg.TableID))
		if err != nil {
			return err
		}
		result.OldTable = result.Table

		if result.Table.Occupied+int32(arg.NewEntourage)+1 > result.Table.Size {
			return InsufficientTableSizeErr(int(result.Table.ID))
		}

		arrivalSQL, err := q.CreateArrival(ctx, CreateArrivalParams{
			GuestID:   int32(arg.UserID),
			TableID:   int32(arg.TableID),
			PartySize: int32(arg.NewEntourage) + 1,
		})

		if err != nil {
			return err
		}

		// Update Original guest record with new entourage value
		err = q.UpdateGuestArrival(ctx, UpdateGuestArrivalParams{
			ID:        int32(arg.UserID),
			Entourage: int32(arg.NewEntourage),
		})

		if err != nil {
			return err
		}

		err = q.UpdateTable(ctx, UpdateTableParams{
			ID:       result.Table.ID,
			Size:     result.Table.Size,
			Occupied: result.Table.Occupied + int32(arg.NewEntourage) + 1,
		})

		if err != nil {
			return err
		}

		// Reading the Arrival object here due to the no Returning property of MySQL
		// This will be added to the table when the transaction is committed,
		// and this can only happen if there are no errors beforehand
		result.Arrival, err = q.getArrivalFromSQLQuery(arrivalSQL)
		if err != nil {
			return err
		}

		result.Guest, err = q.GetGuest(ctx, int32(arg.UserID))
		if err != nil {
			return err
		}

		result.Table, err = q.GetTable(ctx, int32(arg.TableID))
		if err != nil {
			return err
		}

		return nil
	})
	return result, err
}

// DeleteGuestTx deletes a guest from the guests table while also freeing up their table space
// if they've already arrived
func (store *SQLStore) DeleteGuestTx(ctx context.Context, id int32) error {

	err := store.execTx(ctx, func(q *Queries) error {

		guest, err := q.GetGuest(ctx, id)
		if err != nil {
			return err
		}

		err = q.DeleteGuest(ctx, id)
		if err != nil {
			return err
		}

		table, errTable := q.GetTableForUpdate(ctx, guest.TableID)
		// Swallow any errors here as it's acceptable that a guest hasn't yet arrived successfully at a table
		if errTable == nil {
			err = q.UpdateTable(ctx, UpdateTableParams{
				ID:       table.ID,
				Size:     table.Size,
				Occupied: table.Occupied - (guest.Entourage + 1),
			})

			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}
