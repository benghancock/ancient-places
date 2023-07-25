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
	counts := queryCountryCounts(db)
	for countryName, placeCount := range counts {
		fmt.Printf("%s\t%d\n", countryName, placeCount)
	}
}

func queryCountryCounts(db *sql.DB) map[string]int {
	var counts = make(map[string]int)
	q := `
		SELECT
			COALESCE(country_name, '(unknown)'),
			COUNT(DISTINCT(place_id)) AS place_count
		FROM countries_places
		GROUP BY country_name;
`
	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var countryName string
		var placeCount int
		if err := rows.Scan(&countryName, &placeCount); err != nil {
			log.Fatal(err)
		}
		counts[countryName] = placeCount
	}
	return counts
}
