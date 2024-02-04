package api

import (
	"net/http"

	db "github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc"
	"github.com/gin-gonic/gin"
)

type createTableRequest struct {
	Size int32 `json:"size" binding:"required,min=1"`
}

// createTable godoc
// @Summary Creates a table according to the table size.
// @Description Executes a POST request adding the table object to the db..
// @Accept json
// @Produce json
// @Param    size     body      int     true  "Table Size - minimum value is 1"
// @Success 200 {object} sql.Result
// @Failure 400 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /tables/ [post]
func (server *Server) createTable(ctx *gin.Context) {
	var req createTableRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.CreateTableParams{
		Size:     req.Size,
		Occupied: 0,
	}

	table, err := server.store.CreateTable(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, table)
}

type getTablesRequest struct {
	PageID   int32 `form:"page_id" binding:"required,min=1"`
	PageSize int32 `form:"page_size" binding:"required,min=5,max=10"`
}

// getTables godoc
// @Summary returns all tables
// @Description Fetches an array of table object ([]Table), the requests are paginated with a minimum page_id of 1 and page_size of 5-20. Running a make test will generate some default data via the mysql unit tests.
// @Accept json
// @Produce json
// @Param        page_id   query      int  true  "Page ID"
// @Param        page_size   query      int  true  "Page Size"
// @Success 200 {object} []db.Table
// @Failure 400 {object} httputil.HTTPError
// @Failure 500 {object} httputil.HTTPError
// @Router /tables/ [get]
func (server *Server) getTables(ctx *gin.Context) {
	var req getTablesRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	arg := db.GetTablesParams{
		Limit:  req.PageSize,
		Offset: (req.PageID - 1) * req.PageSize,
	}

	tables, err := server.store.GetTables(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, tables)
}

// getEmptySeats godoc
// @Summary Gets all the empty seats
// @Description The empty seats are calculated from the difference between the Size and Occupied values in the table.
// @Accept json
// @Produce json
// @Success 200 {object} int
// @Failure 500 {object} httputil.HTTPError
// @Router /seats_empty [get]
func (server *Server) getEmptySeats(ctx *gin.Context) {
	count, err := server.store.GetEmptySeats(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}

	ctx.JSON(http.StatusOK, count)
}
