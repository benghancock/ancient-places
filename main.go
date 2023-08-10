// package main connects to the archaia database and does the work
// this is a work in progress ...
package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
	"os"
    "net/http"

    "github.com/labstack/echo/v4"
)

var db *sql.DB

type SearchResult struct {
	Count	int	`json:count`
	Results	[]PleiadesPlace `json:results`
}

type PleiadesPlace struct {
	Name	string	`json:name`
	Country	string	`json:country`
	PlaceType	string	`json:placeType`
	Description	string	`json:description`
}

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

	e := echo.New()
	e.GET("/search", searchPlaces)
	e.Logger.Fatal(e.Start(":1323"))
}

func searchPlaces(c echo.Context) error {
	result := new(SearchResult)
	country := c.QueryParam("country")
	places := queryCountryPlaces(db, country)
	result.Count = len(places)
	result.Results = places
	return c.JSON(http.StatusOK, result)
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

// queryCountryPlaces returns a slice of matching PleiadesPlaces
func queryCountryPlaces (db *sql.DB, name string) []PleiadesPlace {
	var matchPlaces []PleiadesPlace
	q := `
		SELECT place_name, country_name, place_type, descrip
		FROM countries_places
		WHERE country_name ILIKE '%' || $1 || '%';
	`
	rows, err := db.Query(q, name)
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var place PleiadesPlace
		if err := rows.Scan(&place.name, &place.country, &place.placeType, &place.description); err != nil {
			log.Fatal(err)
		}
		matchPlaces = append(matchPlaces, place)
	}
	return matchPlaces
}
