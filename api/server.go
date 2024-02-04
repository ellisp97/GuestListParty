package api

import (
	db "github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc"
	docs "github.com/ellisp97/BE_Task_Oct20/golang/docs"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/gin-swagger/swaggerFiles"
)

// Server serves HTTP requests for the guestlist service
type Server struct {
	store  db.Store
	router *gin.Engine
}

// NewSever implements a new HTTP Server and sets up routing
func NewServer(store db.Store) *Server {
	server := &Server{store: store}
	router := gin.Default()

	router.POST("/guest_list/:name", server.createGuest)
	router.PUT("/guests/:name", server.arriveGuest)
	router.GET("/guests/:name", server.getGuestFromName)
	router.GET("/guest_list", server.getGuests)
	router.GET("/guests", server.getArrivedGuests)
	router.GET("/seats_empty", server.getEmptySeats)
	router.GET("/tables", server.getTables)
	router.POST("/tables", server.createTable)
	router.DELETE("/guest/:name", server.deleteGuest)

	// Set up documentation
	docs.SwaggerInfo.BasePath = "/"
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("http://localhost:3000/swagger/doc.json"),
		ginSwagger.DefaultModelsExpandDepth(-1)))
	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

// Wrapper for gin errors
func errorResponse(err error) gin.H {
	return gin.H{"error": err.Error()}
}
