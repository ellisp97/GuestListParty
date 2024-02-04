package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ellisp97/BE_Task_Oct20/golang/util"

	"github.com/stretchr/testify/require"
)

func createRandomGuest(t *testing.T, tableID int32) Guest {
	arg := CreateGuestParams{
		GuestName:   util.RandomGuestName(),
		Entourage:   util.RandomGuestSize(),
		TableID:     tableID,
		ArrivalTime: util.RandomGuestArrivalTime(),
	}

	guestSQL, err := testQueries.CreateGuest(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, guestSQL)

	guest, err := testQueries.getGuestFromSQLQuery(guestSQL)
	require.NoError(t, err)
	require.NotEmpty(t, guest)

	require.Equal(t, arg.GuestName, guest.GuestName)
	require.NotZero(t, guest.ID)
	require.GreaterOrEqual(t, int(guest.Entourage), 0)

	return guest
}

func TestCreateGuest(t *testing.T) {
	table := createRandomTable(t)
	createRandomGuest(t, table.ID)
}

func TestGetGuest(t *testing.T) {
	table := createRandomTable(t)
	guest1 := createRandomGuest(t, table.ID)
	guest2, err := testQueries.GetGuest(context.Background(), guest1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, guest1)

	require.Equal(t, guest1.ID, guest2.ID)
	require.Equal(t, guest1.GuestName, guest2.GuestName)
	require.Equal(t, guest1.Entourage, guest2.Entourage)
	require.Equal(t, guest1.TableID, guest2.TableID)
	require.WithinDuration(t, guest1.ArrivalTime.Time, guest2.ArrivalTime.Time, 2*time.Second)
	require.WithinDuration(t, guest1.CreatedAt.Time, guest2.CreatedAt.Time, 2*time.Second)
}

func TestUpdateGuest(t *testing.T) {
	table := createRandomTable(t)
	guest1 := createRandomGuest(t, table.ID)

	arg := UpdateGuestArrivalParams{
		ID:        guest1.ID,
		Entourage: util.RandomGuestSize(),
	}

	err := testQueries.UpdateGuestArrival(context.Background(), arg)
	require.NoError(t, err)

	guest2, err := testQueries.GetGuest(context.Background(), guest1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, guest2)

	require.Equal(t, guest1.ID, guest2.ID)
	require.Equal(t, guest1.GuestName, guest2.GuestName)
	require.Equal(t, arg.Entourage, guest2.Entourage)
	require.Equal(t, guest1.TableID, guest2.TableID)
	require.WithinDuration(t, guest1.ArrivalTime.Time, guest2.ArrivalTime.Time, 2*time.Second)
	require.WithinDuration(t, guest1.CreatedAt.Time, guest2.CreatedAt.Time, 2*time.Second)
}

func TestDeleteGuest(t *testing.T) {
	table := createRandomTable(t)
	guest1 := createRandomGuest(t, table.ID)

	err := testQueries.DeleteGuest(context.Background(), guest1.ID)
	require.NoError(t, err)

	guest2, err := testQueries.GetGuest(context.Background(), guest1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, guest2)
}

func TestListGuests(t *testing.T) {
	table := createRandomTable(t)

	for i := 0; i < 10; i++ {
		createRandomGuest(t, table.ID)
	}

	// skip first 5 records, return next 5
	arg := GetGuestsParams{
		Limit:  5,
		Offset: 5,
	}

	guests, err := testQueries.GetGuests(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, guests, 5)

	for _, guest := range guests {
		require.NotEmpty(t, guest)
	}
}
