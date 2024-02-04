package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/ellisp97/BE_Task_Oct20/golang/util"

	"github.com/stretchr/testify/require"
)

func TestAssignTableTx(t *testing.T) {
	store := NewStore(testDB)

	n := 6

	errs := make(chan error)
	results := make(chan AssignTableTxResult)

	variations := 3
	var differentTables []Table

	for i := 0; i < variations; i++ {
		differentTables = append(differentTables, createRandomTable(t))
	}

	for i := 0; i < n; i++ {

		table := differentTables[util.RandomInt(0, int32(variations)-1)]
		guest := createRandomGuest(t, table.ID)

		txName := fmt.Sprintf("tx %d", i+1)
		go func() {
			ctx := context.WithValue(context.Background(), txKey, txName)
			result, err := store.AssignTableTx(ctx, AssignTableTxParams{
				UserID:       int64(guest.ID),
				TableID:      int64(table.ID),
				NewEntourage: int64(guest.Entourage),
			})

			errs <- err
			results <- result
		}()
	}

	for i := 0; i < n; i++ {
		err := <-errs
		result := <-results

		// Result should be non empty even if an InsufficientTableSizeErr is returned
		require.NotEmpty(t, result)

		// Same condition with the guest and table requires below
		guest := result.Guest
		require.NotEmpty(t, guest)

		table := result.Table
		require.NotEmpty(t, table)

		// If the party size was more than the table size we want to check we return the expected error
		// then exit any subsequent checks
		if err != nil && guest.Entourage+table.Occupied+1 > table.Size {
			require.Error(t, err)
			require.EqualError(t, InsufficientTableSizeErr(int(table.ID)), err.Error())
			continue
		}

		// The only contained error we're allowing is the InsufficientSizeErr everything else is unexpected
		require.NoError(t, err)

		arrival := result.Arrival

		require.NotEmpty(t, arrival)
		require.NotZero(t, arrival.ID)
		require.Equal(t, arrival.GuestID, guest.ID)
		require.Equal(t, arrival.TableID, table.ID)
		require.Equal(t, arrival.PartySize, guest.Entourage+1)

		_, err = store.GetArrival(context.Background(), arrival.ID)
		require.NoError(t, err)

		oldTable := result.OldTable
		require.NotEmpty(t, oldTable)
		require.Equal(t, guest.Entourage, arrival.PartySize-1)
		require.Equal(t, table.Occupied, oldTable.Occupied+arrival.PartySize)
	}
}

func TestDeleteGuestTx(t *testing.T) {
	store := NewStore(testDB)

	// Transaction should at least pass default delete cases before checking transaction states
	TestDeleteGuest(t)

	table := createRandomTable(t)

	arg := CreateGuestParams{
		GuestName:   util.RandomGuestName(),
		Entourage:   table.Size - 1,
		TableID:     table.ID,
		ArrivalTime: util.RandomGuestArrivalTime(),
	}

	guestSQL, err := testQueries.CreateGuest(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, guestSQL)

	guest, err := testQueries.getGuestFromSQLQuery(guestSQL)
	require.NoError(t, err)
	require.NotEmpty(t, guest)

	assignTableTxResult, err := store.AssignTableTx(context.Background(), AssignTableTxParams{
		UserID:       int64(guest.ID),
		TableID:      int64(guest.TableID),
		NewEntourage: int64(guest.Entourage),
	})
	require.NoError(t, err)
	require.NotEmpty(t, assignTableTxResult)

	arrival := assignTableTxResult.Arrival
	require.NotEmpty(t, arrival)

	newTable := assignTableTxResult.Table
	require.NotEmpty(t, newTable)
	require.Equal(t, newTable.Size, arrival.PartySize)

	err = store.DeleteGuestTx(context.Background(), guest.ID)
	require.NoError(t, err)

	guest2, err := testQueries.GetGuest(context.Background(), guest.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, guest2)

	// Finally check if the Table Occupied size has been decreased following the guests deletion
	table, err = testQueries.GetTable(context.Background(), table.ID)
	require.NoError(t, err)
	require.Equal(t, int(table.Occupied), 0)
}
