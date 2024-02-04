package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/ellisp97/BE_Task_Oct20/golang/util"

	"github.com/stretchr/testify/require"
)

func createRandomTable(t *testing.T) Table {
	arg := CreateTableParams{
		Size:     util.RandomTableSize(),
		Occupied: 0,
	}

	tableSQL, err := testQueries.CreateTable(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, tableSQL)

	table, err := testQueries.getTableFromSQLQuery(tableSQL)
	require.NoError(t, err)
	require.NotEmpty(t, tableSQL)

	require.Equal(t, arg.Size, table.Size)
	require.NotZero(t, table.ID)
	require.Zero(t, table.Occupied)

	return table
}

func TestCreateTable(t *testing.T) {
	createRandomTable(t)
}

func TestGetTable(t *testing.T) {
	table1 := createRandomTable(t)
	table2, err := testQueries.GetTable(context.Background(), table1.ID)

	require.NoError(t, err)
	require.NotEmpty(t, table1)

	require.Equal(t, table1.ID, table2.ID)
	require.Equal(t, table1.Size, table2.Size)
	require.Equal(t, table1.Occupied, table2.Occupied)
	require.WithinDuration(t, table1.CreatedAt.Time, table2.CreatedAt.Time, 2*time.Second)
}

func TestUpdateTableSize(t *testing.T) {
	table1 := createRandomTable(t)

	arg := UpdateTableParams{
		ID:       table1.ID,
		Size:     util.RandomTableSize(),
		Occupied: 0,
	}

	err := testQueries.UpdateTable(context.Background(), arg)
	require.NoError(t, err)

	table2, err := testQueries.GetTable(context.Background(), table1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, table2)

	require.Equal(t, table1.ID, table2.ID)
	require.Equal(t, arg.Size, table2.Size)
	require.Equal(t, table1.Occupied, table2.Occupied)
	require.WithinDuration(t, table1.CreatedAt.Time, table2.CreatedAt.Time, 2*time.Second)
}

func TestDeleteTable(t *testing.T) {
	table1 := createRandomTable(t)

	err := testQueries.DeleteTable(context.Background(), table1.ID)
	require.NoError(t, err)

	table2, err := testQueries.GetTable(context.Background(), table1.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, table2)
}

func TestListTables(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomTable(t)
	}

	// skip first 5 records, return next 5
	arg := GetTablesParams{
		Limit:  5,
		Offset: 5,
	}

	tables, err := testQueries.GetTables(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, tables, 5)

	for _, table := range tables {
		require.NotEmpty(t, table)
	}
}

func TestGetEmptySeats(t *testing.T) {
	count, err := testQueries.GetEmptySeats(context.Background())
	require.NoError(t, err)
	require.NotNil(t, count)

	tableSize := 0
	tableOccupied := 0
	tables, err := testQueries.GetTables(context.Background(), GetTablesParams{
		Limit:  1000,
		Offset: 0,
	})
	require.NoError(t, err)

	for _, table := range tables {
		tableSize += int(table.Size)
		tableOccupied += int(table.Occupied)
	}

	require.NoError(t, err)
	require.Equal(t, count, int32(tableSize-tableOccupied))
}
