// package main connects to the archaia database and does the work
// this is a work in progress ...
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
	"sort"
)

var db *sql.DB

func main() {
	dbUser := os.Getenv("DBUSER")
	dbPass := os.Getenv("DBPASSWORD")
	dsn := fmt.Sprintf(
		"user=%s password=%s dbname=archaia sslmode=disable\n",
		dbUser,
		dbPass)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	log.Println("Connected to archaia!")
	counts := queryCountryCounts(db)
	keys := make([]string, 0, len(counts))
	for k := range counts {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return counts[keys[i]] > counts[keys[j]]
	})
	for _, k := range keys {
		fmt.Printf("%s\t%d\n", k, counts[k])
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
