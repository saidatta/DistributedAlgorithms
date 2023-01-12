package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"sync"
)

func main() {
	// Connect to the database
	db, err := sql.Open("postgres", "user=postgres password=mypassword dbname=mydatabase sslmode=disable")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Create a wait group to synchronize the goroutines
	var wg sync.WaitGroup

	// Start a goroutine for each transaction
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			// Start a transaction
			tx, err := db.Begin()
			if err != nil {
				panic(err)
			}

			// Execute a statement that will be part of the transaction
			_, err = tx.Exec(fmt.Sprintf("INSERT INTO users (name) VALUES ('user%d')", i))
			if err != nil {
				panic(err)
			}

			// Commit the transaction
			err = tx.Commit()
			if err != nil {
				panic(err)
			}
		}(i)
	}

	// Wait for all goroutines to finish
	wg.Wait()
}
