package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	mockdb "github.com/ellisp97/BE_Task_Oct20/golang/db/mock"
	db "github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc"
	"github.com/ellisp97/BE_Task_Oct20/golang/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetEmptySeats(t *testing.T) {
	table := randomTable()

	testCases := []struct {
		name          string
		tableID       int32
		count         int32
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:    "OK",
			tableID: table.ID,
			count:   table.Size - table.Occupied,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmptySeats(gomock.Any()).
					Times(1).
					Return(table.Size-table.Occupied, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireEmptySeatsCorrect(t, recorder.Body, int(table.Size-table.Occupied))
			},
		}, {
			name:    "InternalError",
			tableID: table.ID,
			count:   table.Size - table.Occupied,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetEmptySeats(gomock.Any()).
					Times(1).
					Return(int32(0), sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {

		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			url := "/seats_empty"
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetTablesAPI(t *testing.T) {
	n := 5
	tables := make([]db.Table, n)
	for i := 0; i < n; i++ {
		tables[i] = randomTable()
	}

	type Query struct {
		pageID   int
		pageSize int
	}

	testCases := []struct {
		name          string
		query         Query
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetTablesParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					GetTables(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(tables, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchTables(t, recorder.Body, tables)
			},
		},
		{
			name: "BadRequest",
			query: Query{
				pageID:   -1,
				pageSize: 100,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetTablesParams{
					Limit:  -1,
					Offset: 100,
				}

				store.EXPECT().
					GetTables(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		}, {
			name: "InternalError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetTablesParams{
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().
					GetTables(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(make([]db.Table, 0), sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			req, err := http.NewRequest(http.MethodGet, "/tables", nil)
			require.NoError(t, err)
			params := req.URL.Query()
			params.Add("page_id", fmt.Sprintf("%d", tc.query.pageID))
			params.Add("page_size", fmt.Sprintf("%d", tc.query.pageSize))
			req.URL.RawQuery = params.Encode()

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func TestCreateTableAPI(t *testing.T) {
	table := randomTable()

	testCases := []struct {
		name          string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"size": table.Size,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateTable(gomock.Any(), db.CreateTableParams{Size: table.Size, Occupied: 0}).Times(1)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name: "InvalidNameURI",
			body: nil,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateTable(gomock.Any(), db.CreateTableParams{Size: table.Size, Occupied: 0}).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		}, {
			name: "InternalError",
			body: gin.H{
				"size": table.Size,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().CreateTable(gomock.Any(), db.CreateTableParams{
					Size:     table.Size,
					Occupied: 0,
				}).Times(1).Return(nil, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
	}

	for i := range testCases {

		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			store := mockdb.NewMockStore(controller)
			tc.buildStubs(store)

			server := NewServer(store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/tables", bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

// randomTable returns random test copy of table object to mock
func randomTable() db.Table {
	size := util.RandomTableSize()
	return db.Table{
		ID:       util.RandomInt(1, 1000),
		Size:     size,
		Occupied: size - util.RandomInt(1, size),
	}
}

// requireEmptySeatsCorrect requires mock return object to be equal to responder body when parsed
func requireEmptySeatsCorrect(t *testing.T, body *bytes.Buffer, emptySeats int) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var emptySeatsFetched int
	err = json.Unmarshal(data, &emptySeatsFetched)
	require.NoError(t, err)
	require.Equal(t, emptySeatsFetched, emptySeats)
}

// requireBodyMatchTables requires mock returned table array to be equal to the expected value
func requireBodyMatchTables(t *testing.T, body *bytes.Buffer, guests []db.Table) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var tablesFetched []db.Table
	err = json.Unmarshal(data, &tablesFetched)
	require.NoError(t, err)
	require.Equal(t, guests, tablesFetched)
}
