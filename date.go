package notionapi

import (
	"fmt"
	"strings"
	"time"
)

// Date represents a date
type Date struct {
	// "MMM DD, YYYY", "MM/DD/YYYY", "DD/MM/YYYY", "YYYY/MM/DD", "relative"
	DateFormat string    `json:"date_format"`
	Reminder   *Reminder `json:"reminder,omitempty"`
	// "2018-07-12"
	StartDate string `json:"start_date"`
	// "09:00"
	StartTime string `json:"start_time,omitempty"`
	// "2018-07-12"
	EndDate string `json:"end_date,omitempty"`
	// "09:00"
	EndTime string `json:"end_time,omitempty"`
	// "America/Los_Angeles"
	TimeZone *string `json:"time_zone,omitempty"`
	// "H:mm" for 24hr, not given for 12hr
	TimeFormat string `json:"time_format,omitempty"`
	// "date", "datetime", "datetimerange", "daterange"
	Type string `json:"type"`
}

// Reminder represents a date reminder
type Reminder struct {
	Time  string `json:"time"` // e.g. "09:00"
	Unit  string `json:"unit"` // e.g. "day"
	Value int64  `json:"value"`
}

// parseNotionDateTime parses date and time as sent in JSON by notion
// server and returns time.Time
// date is sent in "2019-04-09" format
// time is optional and sent in "00:35" format
func parseNotionDateTime(date string, t string) time.Time {
	s := date
	fmt := "2006-01-02"
	if t != "" {
		fmt += " 15:04"
		s += " " + t
	}
	dt, err := time.Parse(fmt, s)
	if err != nil {
		MaybePanic("time.Parse('%s', '%s') failed with %s", fmt, s, err)
	}
	return dt
}

// convertNotionTimeFormatToGoFormat converts a date format sent from Notion
// server, e.g. "MMM DD, YYYY" to Go time format like "02 01, 2006"
// YYYY is numeric year => 2006 in Go
// MM is numeric month => 01 in Go
// DD is numeric day => 02 in Go
// MMM is named month => Jan in Go
func convertNotionTimeFormatToGoFormat(d *Date, withTime bool) string {
	format := d.DateFormat
	// we don't support relative time, so use this fixed format
	if format == "relative" || format == "" {
		format = "MMM DD, YYYY"
	}
	s := format
	s = strings.Replace(s, "MMM", "Jan", -1)
	s = strings.Replace(s, "MM", "01", -1)
	s = strings.Replace(s, "DD", "02", -1)
	s = strings.Replace(s, "YYYY", "2006", -1)
	if withTime {
		// this is 24 hr format
		if d.TimeFormat == "H:mm" {
			s += " 15:04"
		} else {
			// use 12 hr format
			s += " 3:04 PM"
		}
	}
	return s
}

// formatDateTime formats date/time from Notion canonical format to
// user-requested format
func formatDateTime(d *Date, date string, t string) string {
	withTime := t != ""
	dt := parseNotionDateTime(date, t)
	goFormat := convertNotionTimeFormatToGoFormat(d, withTime)
	s := dt.Format(goFormat)
	// TODO: this is a lousy way of doing it
	for i := 0; i <= 9; i++ {
		toReplace := fmt.Sprintf("0%d:", i)
		replacement := fmt.Sprintf("%d:", i)
		s = strings.Replace(s, toReplace, replacement, 1)
	}
	// TODO: also timezone
	return s
}

// FormatDate provides default formatting for Date
// TODO: add time zone, maybe
func FormatDate(d *Date) string {
	s := formatDateTime(d, d.StartDate, d.StartTime)
	if strings.Contains(d.Type, "range") {
		s2 := formatDateTime(d, d.EndDate, d.EndTime)
		s += " → " + s2
	}
	return s
}
