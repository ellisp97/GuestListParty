package main

import (
	"database/sql"
	"log"

	"github.com/ellisp97/BE_Task_Oct20/golang/api"
	db "github.com/ellisp97/BE_Task_Oct20/golang/db/sqlc"
	"github.com/ellisp97/BE_Task_Oct20/golang/util"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/golang/mock/mockgen/model"
)

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("Cannot load config file:", err)
	}

	connection, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("Cannot connect to the mysql database: ", err)
	}

	store := db.NewStore(connection)
	server := api.NewServer(store)

	err = server.Start(config.SeverAddress)
	if err != nil {
		log.Fatal("Server failed to start")
	}
}
