package main

import (
	"log"
	"os"

	"github.com/LanaAntonova/go-final-proj/pkg/db"
	"github.com/LanaAntonova/go-final-proj/pkg/server"
)

func main() {
	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}

	err := db.Init(dbFile)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Сервер запущен")
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
