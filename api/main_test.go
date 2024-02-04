package api

import (
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMain(m *testing.M) {
	// Set gin to test mode to reduce logs
	gin.SetMode(gin.TestMode)
	os.Exit(m.Run())
}
