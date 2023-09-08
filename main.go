// package main connPageects to the archaia database and serves a search page
// over HTTP allowing a user search for places by country name
package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
	"golang.org/x/text/message"
)

var db *sql.DB

const pageSize = 200

type config struct {
	DSN string `json:"dsn"`
}

type countryListing struct {
	Name       string
	PlaceCount int
}

type pageData struct {
	PageTitle string
	Data      interface{}
}

type resultsPage struct {
	SearchString   string          `json:"searchString"`
	Count          int             `json:"count"`
	PageNo         int             `json:"pageNo"`
	NextPage       int             `json:"nextPage"`
	DispThisPage   int             `json:"dispThisPage"`
	DispTotalPages int             `json:"dispTotalpages"`
	FmtCount       string          `json:"fmtCount"`
	MoreResults    bool            `json:"hasMoreResults"`
	Results        []pleiadesPlace `json:"results"`
}

type pleiadesPlace struct {
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
	conf := config{}
	decoder.Decode(&conf)
	dsn := conf.DSN

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e := echo.New()
	e.Debug = true
	e.Renderer = t
	e.HTTPErrorHandler = customHTTPErrorHandler

	e.Static("/static", "public/assets")
	e.File("/about.html", "public/about.html")

	homeHandler := func(c echo.Context) error {
		return buildHomepage(c, db)
	}
	e.GET("/", homeHandler)

	searchHandler := func(c echo.Context) error {
		return searchResults(c, db)
	}
	e.GET("/search", searchHandler)

	e.Logger.Fatal(e.Start(":1323"))
}

// serve error pages - source: https://echo.labstack.com/docs/error-handling
func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError // default?
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(err)
	pd := new(pageData)
	pd.PageTitle = strconv.Itoa(code)
	if err := c.Render(code, "error", pd); err != nil {
		c.Logger().Error(err)
	}
}

func buildHomepage(c echo.Context, db *sql.DB) error {
	countryListings := queryCountries(db)
	pd := new(pageData)
	pd.PageTitle = "Home"
	pd.Data = countryListings
	return c.Render(http.StatusOK, "home", pd)
}

// searchResults builds the country search results page
func searchResults(c echo.Context, db *sql.DB) error {
	result := new(resultsPage)
	country := c.QueryParam("country")
	page, err := strconv.Atoi(c.QueryParam("page"))
	if err != nil {
		page = 0
	}

	matchCount := queryMatchCount(db, country)
	places := queryCountryPlaces(db, country, page)
	var hasMoreResults bool
	if ((page + 1) * pageSize) > matchCount {
		hasMoreResults = false
	} else {
		hasMoreResults = true
	}

	totalPages := math.Ceil(float64(matchCount / pageSize))
	fmtPrinter := message.NewPrinter(message.MatchLanguage("en"))

	result.SearchString = country
	result.Count = matchCount
	result.PageNo = page
	result.NextPage = page + 1

	// Data for display on the front end:
	// Show the total counts with a thousands separator.
	// Zero-indexed paging can be confusing.
	result.FmtCount = fmtPrinter.Sprintf("%d", matchCount)
	result.DispThisPage = page + 1
	result.DispTotalPages = int(totalPages) + 1

	result.MoreResults = hasMoreResults
	result.Results = places

	pd := new(pageData)
	pd.PageTitle = "Search Results"
	pd.Data = result

	return c.Render(http.StatusOK, "results", pd)
}

// queryCountries returns a slice of all countries in the db
func queryCountries(db *sql.DB) []countryListing {
	var countries []countryListing
	q := `
		SELECT country_name, COUNT(place_name)
		FROM countries_places
		WHERE country_name IS NOT NULL
		GROUP BY country_name
		ORDER BY country_name ASC
	`
	rows, err := db.Query(q)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var country countryListing
		if err := rows.Scan(
			&country.Name,
			&country.PlaceCount,
		); err != nil {
			log.Fatal(err)
		}
		countries = append(countries, country)
	}
	return countries
}

// queryMatchCount returns a count of matching place results
func queryMatchCount(db *sql.DB, name string) int {
	var count int
	q := `
		SELECT COUNT(place_name)
		FROM countries_places
		WHERE country_name ILIKE '%' || $1 || '%'
	`
	row := db.QueryRow(q, name)
	row.Scan(&count)
	return count
}

// queryCountryPlaces returns a slice of matching PleiadesPlaces
func queryCountryPlaces(db *sql.DB, name string, page int) []pleiadesPlace {
	var matchPlaces []pleiadesPlace

	offset := pageSize * page
	q := `
		SELECT
			place_name,
			country_name,
			place_type,
			COALESCE(descrip, ''),
			pleiades_uri
		FROM countries_places
		WHERE country_name ILIKE '%' || $1 || '%'
		ORDER BY place_name ASC
		LIMIT $2 OFFSET $3;
	`
	rows, err := db.Query(q, name, pageSize, offset)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var place pleiadesPlace
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
