package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mockdb "github.com/ellisp97/BE_Task_Oct20/golang/db/mock"
	db "github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc"
	"github.com/ellisp97/BE_Task_Oct20/golang/util"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestGetArrivedGuests(t *testing.T) {
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

				arg := db.GetArrivedGuestsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					GetArrivedGuests(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(guests, nil)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGuests(t, recorder.Body, guests)
			},
		},
		{
			name: "InternalServerError",
			query: Query{
				pageID:   1,
				pageSize: n,
			},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetArrivedGuestsParams{
					Limit:  int32(n),
					Offset: 0,
				}

				store.EXPECT().
					GetArrivedGuests(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(make([]db.Guest, 0), sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		}, {
			name:  "InvalidPagination",
			query: Query{},
			buildStubs: func(store *mockdb.MockStore) {

				arg := db.GetArrivedGuestsParams{
					Limit:  int32(n),
					Offset: 0,
				}
				store.EXPECT().
					GetArrivedGuests(gomock.Any(), gomock.Eq(arg)).
					Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			req, err := http.NewRequest(http.MethodGet, "/guests", nil)
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

func TestPutArrivedGuest(t *testing.T) {
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
			},
			buildStubs: func(store *mockdb.MockStore) {
				first := store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).Times(1).Return(guest, nil)
				second := store.EXPECT().AssignTableTx(gomock.Any(), gomock.Eq(db.AssignTableTxParams{
					UserID:       int64(guest.ID),
					TableID:      int64(guest.TableID),
					NewEntourage: int64(guest.Entourage),
				})).
					Times(1).
					Return(createAssignTxTableResult(guest, table, int(guest.Entourage)), nil)
				gomock.InOrder(first, second)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchGuestName(t, recorder.Body, guest.GuestName)
			},
		},
		{
			name:      "NewEntourageTooBig",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": guest.Entourage + table.Size, // enforce it's too big for the table
			},
			buildStubs: func(store *mockdb.MockStore) {
				first := store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).Times(1).Return(guest, nil)
				second := store.EXPECT().AssignTableTx(gomock.Any(), gomock.Eq(db.AssignTableTxParams{
					UserID:       int64(guest.ID),
					TableID:      int64(guest.TableID),
					NewEntourage: int64(guest.Entourage + table.Size),
				})).
					Times(1).
					Return(db.AssignTableTxResult{}, db.InsufficientTableSizeErr(int(guest.TableID)))
				gomock.InOrder(first, second)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name:      "NoNameFound",
			guestName: "InvalidUser",
			body: gin.H{
				"entourage": guest.Entourage,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq("InvalidUser")).Times(1).Return(db.Guest{}, sql.ErrNoRows)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
		{
			name:      "InternalServerError",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": guest.Entourage,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).Times(1).Return(db.Guest{}, sql.ErrConnDone)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		}, {
			name:      "InternalServerErrorPostGetName",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": guest.Entourage,
			},
			buildStubs: func(store *mockdb.MockStore) {
				first := store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).Times(1).Return(guest, nil)
				second := store.EXPECT().AssignTableTx(gomock.Any(), gomock.Eq(db.AssignTableTxParams{
					UserID:       int64(guest.ID),
					TableID:      int64(guest.TableID),
					NewEntourage: int64(guest.Entourage),
				})).
					Times(1).
					Return(db.AssignTableTxResult{}, sql.ErrConnDone)
				gomock.InOrder(first, second)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		}, {
			name:      "InvalidEntourageValue",
			guestName: guest.GuestName,
			body: gin.H{
				"entourage": -1,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		}, {
			name:      "InvalidNameValue",
			guestName: "-1",
			body: gin.H{
				"entourage": guest.Entourage,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetGuestFromName(gomock.Any(), gomock.Eq(guest.GuestName)).Times(0)
			},
			checkResponse: func(recorder *httptest.ResponseRecorder) {
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

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := fmt.Sprintf("/guests/%s", tc.guestName)
			req, err := http.NewRequest(http.MethodPut, url, bytes.NewReader(data))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, req)
			tc.checkResponse(recorder)
		})
	}
}

func createAssignTxTableResult(guest db.Guest, table db.Table, newEntourage int) db.AssignTableTxResult {
	return db.AssignTableTxResult{
		Arrival:  createArrival(guest.ID, guest.TableID, int32(newEntourage+1)),
		OldTable: table,
		Table:    db.Table{ID: table.ID, Size: table.Size, Occupied: table.Occupied + int32(newEntourage) + 1, CreatedAt: sql.NullTime{Time: time.Now(), Valid: true}},
		Guest:    guest,
	}
}

func createArrival(guestID int32, tableID int32, partySize int32) db.Arrival {
	return db.Arrival{
		ID:        util.RandomInt(1, 1000),
		GuestID:   guestID,
		TableID:   tableID,
		PartySize: partySize,
	}
}
