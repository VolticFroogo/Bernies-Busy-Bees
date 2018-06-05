package main

import (
	"log"

	"github.com/VolticFroogo/Bernies-Busy-Bees/db"
	"github.com/VolticFroogo/Bernies-Busy-Bees/handler"
	"github.com/VolticFroogo/Bernies-Busy-Bees/middleware/myJWT"
)

func main() {
	if err := db.InitDB(); err != nil {
		log.Printf("Error initializing database: %v", err)
		return
	}

	if err := myJWT.InitKeys(); err != nil {
		log.Printf("Error initializing JWT keys: %v", err)
		return
	}

	handler.Start()
}
