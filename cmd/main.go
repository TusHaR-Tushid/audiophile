package main

import (
	"Audiophile/database"
	"Audiophile/server"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	//host := os.Getenv("host")
	//port := os.Getenv("port")
	//databaseName := os.Getenv("databaseName")
	//user := os.Getenv("user")
	//password := os.Getenv("password")
	host := os.Getenv("host")
	port := "5435"
	databaseName := "audiophile"
	user := "postgres"
	password := "1234"

	err := database.ConnectAndMigrate(host, port, databaseName, user, password, database.SSLModeDisable)
	if err != nil {
		logrus.Printf("ConnectAndMigrate: error is:%v", err)
		return
	}
	fmt.Println("connected")
	srv := server.SetupRoutes()
	err = srv.Run(":8080")
	if err != nil {
		logrus.Printf("could not run the server error:%v", err)
		return
	}
}
