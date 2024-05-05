package main

import (
	"embed"
	"fmt"
	"slices"
	"text/template"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

//go:embed templates
var assets embed.FS
var templates *template.Template

func init() {
	templates = template.Must(template.ParseFS(assets, "templates/*.html"))
}

type IndexPayload struct {
	Day     string
	Records []Record
}

func Index(c echo.Context, db *LogitterDB) error {
	// get data from db
	records := db.GetRecords()
	slices.Reverse(records)

	// make groups by day
	payload := []IndexPayload{}
	for _, v := range records {
		currTime := time.Unix(v.Timestamp, 0)
		currDay := fmt.Sprintf("%d-%02d-%02d", currTime.Year(), currTime.Month(), currTime.Day())
		if len(payload) == 0 || payload[len(payload)-1].Day != currDay {
			payload = append(payload, IndexPayload{Day: currDay})
		}
		payload[len(payload)-1].Records = append(payload[len(payload)-1].Records, v)
	}

	// write response to the client
	err := templates.ExecuteTemplate(c.Response().Writer, "index", payload)
	if err != nil {
		panic(err) // TODO
	}
	return nil
}

func ServeFrontend(db *LogitterDB) {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return Index(c, db)
	})

	e.Logger.Fatal(e.Start(":1323"))
}
