package api

import (
	"database/sql"
	"fmt"
	"net/http"

	db "github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/swaggo/swag/example/celler/httputil"
)

type createGuestRequest struct {
	Entourage int32 `json:"entourage" binding:"required"`
	TableID   int32 `json:"table_id" binding:"required,min=0"`
}

// Normally this would go in the above createGuestRequest but to conform to the project
// outline in the README.md it's separated like so
type createGuestRequestURI struct {
	GuestName string `uri:"name" binding:"required,min=5"`
}

// createGuest godoc
// @Summary Creates a guest according to the name, table, and entourage arguments.
// @Description Executes a POST request preceeding the check to see if the table is big enough for the party (1 + entourage).
// @Accept json
// @Produce json
// @Param    name         path      string  true  "Guest Name"
// @Param    entourage    body      int     true  "Entourage"
// @Param    table_id     body      int     true  "Table ID - unique identifier of the table (see getTables)"
// @Success 200 {object} sql.Result
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /guest_list/{name} [post]
func (server *Server) createGuest(ctx *gin.Context) {
	var reqUri createGuestRequestURI
	var reqBody createGuestRequest
	if errUri := ctx.ShouldBindUri(&reqUri); errUri != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errUri))
		return
	}

	if errBody := ctx.ShouldBindJSON(&reqBody); errBody != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(errBody))
		return
	}

	arg := db.CreateGuestParams{
		GuestName:   reqUri.GuestName,
		Entourage:   reqBody.Entourage,
		TableID:     reqBody.TableID,
		ArrivalTime: sql.NullTime{},
	}

	// Must check first the table they provided is big enough
	table, err := server.store.GetTable(ctx, reqBody.TableID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, errorResponse(err))
		return
	}

	if table.Size < (arg.Entourage + 1) {
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("the table size %d is not big enough to hold the capacity of your party %d", table.Size, arg.Entourage+1))
		return
	}

	_, err = server.store.CreateGuest(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, reqUri.GuestName)
}

type getGuestFromNameRequest struct {
	Name string `uri:"name" binding:"required,min=5"`
}

// getGuestFromName godoc
// @Summary returns a guest based on their GuestName value.
// @Description Fetches a guest object (Guest)
// @Accept json
// @Produce json
// @Param    name     path      string  true  "Guest Name"
// @Success 200 {object} db.Guest
// @Failure 400 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /guests/{name} [get]
func (server *Server) getGuestFromName(ctx *gin.Context) {
	var req getGuestFromNameRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	guest, err := server.store.GetGuestFromName(ctx, req.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, guest)
}

type getGuestsRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// @BasePath /

// getGuests godoc
// @Summary returns all guests on the guest_list
// @Description Fetches an array of guest object ([]Guest), the requests are paginated with a minimum page_id of 1 and page_size of 5-20. Running a make test will generate some default data via the mysql unit tests.
// @Accept json
// @Produce json
// @Param        page_id   query      int  true  "Page ID"
// @Param        page_size   query      int  true  "Page Size"
// @Success 200 {object} []db.Guest
// @Failure 400 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /guest_list/ [get]
func (server *Server) getGuests(ctx *gin.Context) {
	var req getGuestsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetGuestsParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	guests, err := server.store.GetGuests(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, guests)
}

// deleteGuest godoc
// @Summary Deletes a guest based on their Guest Name value.
// @Description Checks there is a valid record based on the name value then performs a DELETE action.
// @Accept json
// @Produce json
// @Param        name   path    string  true  "Guest Name"
// @Success 200 {object} []db.Guest
// @Failure 400 {object} httputil.HTTPError
// @Failure 404 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /guest/{name} [delete]
func (server *Server) deleteGuest(ctx *gin.Context) {
	var req getGuestFromNameRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	guest, err := server.store.GetGuestFromName(ctx, req.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	err = server.store.DeleteGuestTx(ctx, guest.ID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, guest.GuestName)
}
