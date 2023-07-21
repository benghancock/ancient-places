// package main connects to the archaia database and does the work
// this is a work in progress ...
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
)

var db *sql.DB

func main() {
	var placeNames []string

	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASSWORD")
	connStr := fmt.Sprintf("user=%s password=%s dbname=archaia sslmode=disable\n", dbUser, dbPass)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected to archaia!")

	rows, err := db.Query("SELECT place_name FROM countries_places WHERE lower(country_name) = 'greece' LIMIT 1;")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var placeName string
		if err := rows.Scan(&placeName); err != nil {
			log.Fatal(err)
		}
		placeNames = append(placeNames, placeName)
	}
	fmt.Println("Places")
	fmt.Println("------")
	for _, name := range placeNames {
		fmt.Println(name)
	}
}
