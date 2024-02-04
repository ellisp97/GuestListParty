package db

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/ellisp97/BE_Task_Oct20/golang/util"

	_ "github.com/go-sql-driver/mysql"
)

var testQueries *Queries
var testDB *sql.DB

func TestMain(m *testing.M) {
	config, err := util.LoadConfig("../..")
	if err != nil {
		log.Fatal("Cannot load config file:", err)
	}

	testDB, err = sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to the mysql database: ", err)
	}

	testQueries = New(testDB)

	os.Exit(m.Run())
}
