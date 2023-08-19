// package main connects to the archaia database and serves a search page
// over HTTP allowing a user search for places by country name
package main

import (
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	"html/template"
	"io"
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
	SearchString string          `json:"searchString"`
	Count        int             `json:"count"`
	Results      []PleiadesPlace `json:"results"`
}

type PleiadesPlace struct {
	Name        string `json:"name"`
	Country     string `json:"country"`
	PlaceType   string `json:"placeType"`
	Description string `json:"description"`
	URI         string `json:"pleiadesURL"`
}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
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

	files := []string{
		"./public/views/base.html",
		"./public/views/results.html",
	}

	t := &Template{
		templates: template.Must(template.ParseFiles(files...)),
	}

	e := echo.New()
	e.Renderer = t

	e.File("/", "public/index.html")
	e.Static("/static", "public/assets")

	e.GET("/search", func(c echo.Context) error {
		result := new(SearchResult)
		country := c.QueryParam("country")
		places := queryCountryPlaces(db, country)

		result.SearchString = country
		result.Count = len(places)
		result.Results = places

		return c.Render(http.StatusOK, "base", result)
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
