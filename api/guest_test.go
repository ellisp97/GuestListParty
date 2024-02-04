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

func TestGetGuestAPI(t *testing.T) {
	guest := randomGuest()

	testCases := []struct {
		name          string
		guestName     string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			guestName: guest.GuestName,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).
					Times(1).
					Return(guest, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGuest(t, recorder.Body, guest)
			},
		}, {

			name:      "NotFound",
			guestName: guest.GuestName,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).
					Times(1).
					Return(db.Guest{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		}, {

			name:      "InternalError",
			guestName: guest.GuestName,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).
					Times(1).
					Return(db.Guest{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidID",
			guestName: "a",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetGuestFromName(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
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

			url := fmt.Sprintf("/guests/%s", tc.guestName)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

type T struct {
	S   int64
	err error
}

func (t T) LastInsertID() (int64, error) {
	return t.S, t.err
}

func (t T) RowsAffected() (int64, error) {
	return t.S, t.err
}

func TestCreateGuestAPI(t *testing.T) {
	table := randomTable()
	table.Occupied = 0 // default it to an empty table

	guest := randomGuest()
	guest.TableID = table.ID
	guest.Entourage = table.Size - 1 // default to always fit

	testCases := []struct {
		name          string
		guestName     string
		body          gin.H
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": guest.Entourage,
				"table_id":  table.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				first := store.EXPECT().GetTable(gomock.Any(), guest.TableID).Times(1).Return(table, nil)
				second := store.EXPECT().CreateGuest(gomock.Any(), db.CreateGuestParams{
					GuestName:   guest.GuestName,
					Entourage:   guest.Entourage,
					TableID:     guest.TableID,
					ArrivalTime: sql.NullTime{Valid: false},
				}).Times(1)
				gomock.InOrder(first, second)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
		{
			name:      "TableTooSmall",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": guest.Entourage + table.Size, // ensure table capacity is overflowed
				"table_id":  table.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTable(gomock.Any(), guest.TableID).Times(1).Return(table, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NotFound",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": guest.Entourage,
				"table_id":  99999,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTable(gomock.Any(), int32(99999)).Times(1).Return(db.Table{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InvalidNameURI",
			guestName: "-1",
			body: gin.H{
				"entourage": guest.Entourage,
				"table_id":  table.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTable(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InvalidBody",
			guestName: guest.GuestName,
			body:      nil,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetTable(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		}, {

			name:      "InternalErrorInCreateGuest",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": guest.Entourage,
				"table_id":  table.ID,
			},
			buildStubs: func(store *mockdb.MockStore) {
				first := store.EXPECT().GetTable(gomock.Any(), guest.TableID).Times(1).Return(table, nil)
				second := store.EXPECT().CreateGuest(gomock.Any(), db.CreateGuestParams{
					GuestName:   guest.GuestName,
					Entourage:   guest.Entourage,
					TableID:     guest.TableID,
					ArrivalTime: sql.NullTime{Valid: false},
				}).Times(1)
				gomock.InOrder(first, second.Return(nil, sql.ErrConnDone))
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

			url := fmt.Sprintf("/guest_list/%s", tc.guestName)
			req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func TestDeleteGuestAPI(t *testing.T) {
	guest := randomGuest()

	testCases := []struct {
		name          string
		guestName     string
		buildStubs    func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name:      "OK",
			guestName: guest.GuestName,
			buildStubs: func(store *mockdb.MockStore) {
				first := store.EXPECT().GetGuestFromName(gomock.Any(), guest.GuestName).Times(1).Return(guest, nil)
				second := store.EXPECT().DeleteGuestTx(gomock.Any(), gomock.Eq(guest.ID)).Times(1).Return(nil)
				gomock.InOrder(first, second)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGuestName(t, recorder.Body, guest.GuestName)
			},
		},
		{
			name:      "NotFound",
			guestName: "UnknownUser",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetGuestFromName(gomock.Any(), "UnknownUser").
					Times(1).
					Return(db.Guest{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalError",
			guestName: guest.GuestName,
			buildStubs: func(store *mockdb.MockStore) {
				first := store.EXPECT().GetGuestFromName(gomock.Any(), guest.GuestName).Times(1).Return(guest, nil)
				second := store.EXPECT().DeleteGuestTx(gomock.Any(), gomock.Eq(guest.ID)).Times(1).Return(sql.ErrConnDone)
				gomock.InOrder(first, second)

			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name:      "InvalidName",
			guestName: "-1",
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Any()).Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "InternalErrorOnFetchName",
			guestName: guest.GuestName,
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).Times(1).Return(db.Guest{}, sql.ErrConnDone)
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

			url := fmt.Sprintf("/guest/%s", tc.guestName)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetGuestsAPI(t *testing.T) {
	n := 5
	guests := make([]db.Guest, n)
	for i := 0; i < n; i++ {
		guests[i] = randomGuest()
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

				arg := db.GetGuestsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					GetGuests(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(guests, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGuests(t, recorder.Body, guests)
			},
		},
		{
			name: "BadRequest",
			query: Query{
				pageID:   -1,
				pageSize: 100,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetGuestsParams{
					Limit:  -1,
					Offset: 100,
				}

				store.EXPECT().
					GetGuests(gomock.Any(), gomock.Eq(arg)).
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

				arg := db.GetGuestsParams{
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().
					GetGuests(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(make([]db.Guest, 0), sql.ErrConnDone)
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

			req, err := http.NewRequest(http.MethodGet, "/guest_list", nil)
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

func randomGuest() db.Guest {
	return db.Guest{
		ID:          util.RandomInt(1, 1000),
		GuestName:   util.RandomGuestName(),
		Entourage:   util.RandomGuestSize(),
		TableID:     util.RandomInt(1, 20),
		ArrivalTime: util.RandomGuestArrivalTime(),
	}
}

// requireBodyMatchGuest requires mock returned guest object to be equal to the expected value
func requireBodyMatchGuest(t *testing.T, body *bytes.Buffer, guest db.Guest) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var guestFetched db.Guest
	err = json.Unmarshal(data, &guestFetched)
	require.NoError(t, err)
	require.Equal(t, guest, guestFetched)
}

// requireBodyMatchGuests requires mock returned guest name to be equal to the expected value
func requireBodyMatchGuestName(t *testing.T, body *bytes.Buffer, pstrName string) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var guestFetched string
	err = json.Unmarshal(data, &guestFetched)
	require.NoError(t, err)
	require.Equal(t, pstrName, guestFetched)
}

// requireBodyMatchGuests requires mock returned guest array to be equal to the expected value
func requireBodyMatchGuests(t *testing.T, body *bytes.Buffer, guests []db.Guest) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)

	var guestsFetched []db.Guest
	err = json.Unmarshal(data, &guestsFetched)
	require.NoError(t, err)
	require.Equal(t, guests, guestsFetched)
}
