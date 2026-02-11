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

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	log.Printf("Server is running on port %04d", server.Port)
	if err := server.Start(); err != nil {
		log.Fatal(err)
	}
}
