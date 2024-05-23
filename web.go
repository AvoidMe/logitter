package main

import (
	"embed"
	"fmt"
	"net/url"
	"slices"
	"strings"
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

func ParseText(text string) string {
	var sb strings.Builder
	arr := strings.Split(text, " ")
	for i, v := range arr {
		if url, err := url.Parse(v); err == nil && (url.Scheme == "http" || url.Scheme == "https") {
			templates.ExecuteTemplate(&sb, "URL", v)
		} else {
			sb.WriteString(v)
		}
		if len(arr) > 1 && i < len(arr)-1 {
			sb.WriteString(" ")
		}
	}
	return sb.String()
}

func IndexPayloadFromDBRecords(records []Record) []IndexPayload {
	// make groups by day
	payload := []IndexPayload{}
	for _, v := range records {
		currTime := time.Unix(v.Timestamp, 0)
		currDay := fmt.Sprintf("%d-%02d-%02d", currTime.Year(), currTime.Month(), currTime.Day())
		if len(payload) == 0 || payload[len(payload)-1].Day != currDay {
			payload = append(payload, IndexPayload{Day: currDay})
		}
		v.Text = ParseText(v.Text)
		payload[len(payload)-1].Records = append(payload[len(payload)-1].Records, v)
	}
	return payload
}

func Index(c echo.Context, db *LogitterDB) error {
	// get data from db
	records := db.GetRecords()
	slices.Reverse(records)

	// group by day
	payload := IndexPayloadFromDBRecords(records)

	// write response to the client
	err := templates.ExecuteTemplate(c.Response().Writer, "Index", payload)
	if err != nil {
		panic(err) // TODO
	}
	return nil
}

func Search(c echo.Context, db *LogitterDB) error {
	// get data from db
	records := db.GetRecordsFilter(c.QueryParam("query"))
	slices.Reverse(records)

	// group by day
	payload := IndexPayloadFromDBRecords(records)

	// write response to the client
	err := templates.ExecuteTemplate(c.Response().Writer, "Items", payload)
	if err != nil {
		panic(err) // TODO
	}
	return nil

}

func ServeFrontend(db *LogitterDB, sigs chan struct{}) {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return Index(c, db)
	})
	e.GET("/search", func(c echo.Context) error {
		return Search(c, db)
	})
	e.GET("/show_ui", func(c echo.Context) error {
		sigs <- struct{}{}
		return nil
	})

	e.Logger.Fatal(e.Start(":7033"))
}
