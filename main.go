// package main connects to the archaia database and does the work
// this is a work in progress ...
package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

var db *sql.DB

type Config struct {
	DSN string `json:"dsn"`
}

type SearchResult struct {
	Count   int             `json:"count"`
	Results []PleiadesPlace `json:"results"`
}

type PleiadesPlace struct {
	Name        string `json:"name"`
	Country     string `json:"country"`
	PlaceType   string `json:"placeType"`
	Description string `json:"description"`
	URI         string `json:"pleiadesURL"`
}

func main() {
	configFile, err := os.Open("conf.json")
	if err != nil {
		log.Fatal("Error loading config file")
	}
	defer configFile.Close()
	decoder := json.NewDecoder(configFile)
	conf := Config{}
	decoder.Decode(&conf)
	dsn := conf.DSN

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	e := echo.New()
	e.File("/", "public/index.html")
	e.GET("/search", func(c echo.Context) error {
		result := new(SearchResult)
		country := c.QueryParam("country")
		places := queryCountryPlaces(db, country)
		result.Count = len(places)
		result.Results = places
		return c.JSON(http.StatusOK, result)
	})
	e.Logger.Fatal(e.Start(":1323"))
}

// queryCountryPlaces returns a slice of matching PleiadesPlaces
func queryCountryPlaces(db *sql.DB, name string) []PleiadesPlace {
	var matchPlaces []PleiadesPlace
	q := `
		SELECT
			place_name,
			country_name,
			place_type,
			COALESCE(descrip, ''),
			pleiades_uri
		FROM countries_places
		WHERE country_name ILIKE '%' || $1 || '%';
	`
	rows, err := db.Query(q, name)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var place PleiadesPlace
		if err := rows.Scan(
			&place.Name,
			&place.Country,
			&place.PlaceType,
			&place.Description,
			&place.URI,
		); err != nil {
			log.Fatal(err)
		}
		matchPlaces = append(matchPlaces, place)
	}
	return matchPlaces
}
