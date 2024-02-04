package api

import (
	"database/sql"
	"net/http"

	db "github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc"
	"github.com/gin-gonic/gin"
)

type arriveGuestRequest struct {
	Entourage int32 `json:"entourage" binding:"required,min=0"`
}

// arriveGuest godoc
// @Summary Arrives the guest into the party
// @Description Performs a PUT action to the arrivals table to record an arrival of the guest and their party.
// @Accept json
// @Produce json
// @Param        name        path       string  true  "Guest Name"
// @Param        entourage   body       int     true  "Entourage (May be different to original)"
// @Success 200 {object} string
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /guests/{name} [put]
func (server *Server) arriveGuest(ctx *gin.Context) {
	var reqName getGuestFromNameRequest
	var reqEntourage arriveGuestRequest
	if errUri := ctx.ShouldBindUri(&reqName); errUri != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errUri))
		return
	}

	if errBody := ctx.ShouldBindJSON(&reqEntourage); errBody != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errBody))
		return
	}

	guest, err := server.store.GetGuestFromName(ctx, reqName.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	arg := db.AssignTableTxParams{
		UserID:       int64(guest.ID),
		TableID:      int64(guest.TableID),
		NewEntourage: int64(reqEntourage.Entourage),
	}

	assignTableResult, err := server.store.AssignTableTx(ctx, arg)
	if err != nil {
		if err.Error() == db.InsufficientTableSizeErr(int(guest.TableID)).Error() {
			ctx.JSON(http.StatusBadRequest, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, assignTableResult.Guest.GuestName)
}

// getArrivedGuests godoc
// @Summary returns all guests already arrived
// @Description Fetches an array of guest object ([]Guest), who have already undergone an arrival event. The requests are paginated with a minimum page_id of 1 and page_size of 5-20. Running a make test will generate some default data via the mysql unit tests.
// @Accept json
// @Produce json
// @Param        page_id     query      int  true  "Page ID"
// @Param        page_size   query      int  true  "Page Size"
// @Success 200 {object} []db.Guest
// @Failure 400 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /guests/ [get]
func (server *Server) getArrivedGuests(ctx *gin.Context) {
	var req getGuestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.GetArrivedGuestsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	guests, err := server.store.GetArrivedGuests(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, guests)
}
