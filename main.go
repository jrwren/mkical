package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"text/template"
	"time"
)

type Event struct {
	DTStart string
	DTEnd   string
	// UID is a consistent globally unique identifier, but if we regen the
	// calendary, then it will be different still, so just use whatever. ;)
	UID         string
	Summary     string
	Description string
	DTStamp     string
	Location    string
	Alarm       bool
}

type Calendar struct {
	Name     string
	Timezone string
	Events   []Event
}

func main() {
	name := flag.String("name", "Mortgage Countdown", "the name of the calendar")
	months := flag.Int("months", 42, "months remaining")
	dom := flag.Int("dom", 0, "day of month for the event, zero means current day of month")
	desctmpl := flag.String("desctmpl", "", "description template")
	location := flag.String("location", "Home", "location of each calendary entry")
	flag.Parse()

	if *dom < 0 || *dom > 28 {
		fmt.Print("invalid dom: dom must be between 0 and 28")
		os.Exit(1)
	}

	if *months < 0 || *months > 500 {
		fmt.Print("invalid months: months must be between 0 and 500")
		os.Exit(2)
	}

	cal := Calendar{
		Name: *name,
	}
	now := time.Now()
	if *dom == 0 {
		*dom = now.Day()
	}
	month := now.Month()
	year := now.Year()
	for i := *months; i > 0; i-- {
		start := time.Date(year, month, *dom, 0, 0, 0, 0, time.Local)
		summary := fmt.Sprintf("%d months remaining on mortgage", i)
		desc := *desctmpl
		if desc == "" {
			desc = summary
		}
		e := Event{
			DTStart:     start.Format("20060102"),
			DTEnd:       start.Format("20060102"),
			Summary:     summary,
			Description: *desctmpl, // TODO expand this template
			DTStamp:     now.Format("20060102T150405Z"),
			Location:    *location,
			UID:         getuid() + "@j.xmtp.net",
		}
		cal.Events = append(cal.Events, e)
		year, month = incYM(year, month)
	}

	ct := `BEGIN:VCALENDAR
NAME:{{.Name}}
X-WR-CALNAME:{{.Name}}
VERSION:2.0
PRODID:-//mkical //mkical//EN
CALSCALE:GREGORION
METHOD:PUBLISH
{{range .Events}}
BEGIN:VEVENT
DTSTART;VALUE=DATE:{{.DTStart}}
DTEND;VALUE=DATE:{{.DTEnd}}
UID:{{.UID}}
SUMMARY:{{.Summary}}
DESCRIPTION:{{.Description}}
LOCATION:{{.Location}}
STATUS:CONFIRMED
DTSTAMP:{{.DTStamp}}
CREATED:{{.DTStamp}}
LAST-MODIFIED:{{.DTStamp}}
TRANSP:OPAQUE
{{if .Alarm}}
BEGIN:VALARM
TRIGGER;VALUE=DURATION:-PT30M
ACTION:DISPLAY
DESCRIPTION:{{.Summary}}
END:VALARM
{{end}}
END:VEVENT
{{end}}
END:VCALENDAR
`
	tmpl, err := template.New("ical").Parse(ct)
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(os.Stdout, cal)
	if err != nil {
		panic(err)
	}
}

func incYM(year int, month time.Month) (int, time.Month) {
	month++
	if month > 12 {
		year++
		month = 1
	}
	return year, month
}

func getuid() string {
	var b [4]byte
	rand.Read(b[:])
	return base64.StdEncoding.EncodeToString(b[:])
}
