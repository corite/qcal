package main

import (
	//"encoding/json"
	"fmt"
	// 	"log"
	"regexp"
	"strings"
	"time"
)

var (
	eventRRuleRegex      = regexp.MustCompile(`RRULE:.*?\n`)
	freqRegex            = regexp.MustCompile(`FREQ=.*?;`)
	eventSummaryRegex    = regexp.MustCompile(`SUMMARY:.*?\n`)
	eventFreqWeeklyRegex = regexp.MustCompile(`RRULE:FREQ=WEEKLY\n`)
	eventFreqYearlyRegex = regexp.MustCompile(`RRULE:FREQ=YEARLY\n`)
)

//  unixtimestamp
const (
	uts = "1136239445"
	//ics date time format
	// Y-m-d H:i:S time format
	YmdHis = "2006-01-02 15:04:05"
)

type Event struct {
	Href        string
	Color       string
	Start       time.Time
	End         time.Time
	TZID        string
	Freq        string
	Summary     string
	Description string
	Location    string
}

func trimField(field, cutset string) string {
	re, _ := regexp.Compile(cutset)
	cutsetRem := re.ReplaceAllString(field, "")
	return strings.TrimRight(cutsetRem, "\r\n")
}

func explodeEvent(eventData *string) (string, string) {
	reEvent, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	Event := reEvent.FindString(*eventData)
	calInfo := reEvent.ReplaceAllString(*eventData, "")
	return Event, calInfo
}

func parseTimeField(fieldName string, eventData string) (time.Time, string) {
	reWholeDay, _ := regexp.Compile(fmt.Sprintf(`%s;VALUE=DATE:.*?\n`, fieldName))
	//re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))?(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))
	// correct regex: .+:(.+)$
	re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.+))?(;VALUE=DATE-TIME)?:(.+?)\n`, fieldName))
	//re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))

	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time
	//var datetime time.Time
	var tzID string
	var format string

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, fmt.Sprintf("%s;VALUE=DATE:", fieldName))
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has start hour and minute
		result := re.FindStringSubmatch(eventData)

		if result == nil || len(result) < 4 {
			return t, tzID
		}

		tzID = result[2]
		dt := result[4]

		//	myLocation, _ := time.LoadLocation("Europe/Berlin")
		if strings.HasSuffix(dt, "Z") {
			// If string end in 'Z', timezone is UTC
			format = "20060102T150405Z"
			time, _ := time.Parse(format, dt)
			t = time.Local()
		} else if tzID != "" {
			format = "20060102T150405"
			location, err := time.LoadLocation(tzID)
			// if tzID not readable use UTC
			if err != nil {
				location, _ = time.LoadLocation("UTC")
			}
			// set foreign timezone
			time, _ := time.ParseInLocation(format, dt, location)
			// convert to local timezone
			//t = time.In(myLocation)
			t = time.Local()
			//fmt.Println(dt)
		} else {
			// Else, consider the timezone is local the parser
			format = "20060102T150405"
			t, _ = time.Parse(format, dt)
		}

	}

	return t, tzID
}

func parseEventStart(eventData *string) (time.Time, string) {
	return parseTimeField("DTSTART", *eventData)
}

func parseEventEnd(eventData *string) (time.Time, string) {
	return parseTimeField("DTEND", *eventData)
}

func parseEventSummary(eventData *string) string {
	re, _ := regexp.Compile(`SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?.*?\n`)
	result := re.FindString(*eventData)
	return trimField(result, `SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?:`)
}

func parseEventDescription(eventData *string) string {
	re, _ := regexp.Compile(`DESCRIPTION:.*?\n(?:\s+.*?\n)*`)

	resultA := re.FindAllString(*eventData, -1)
	result := strings.Join(resultA, ", ")
	//result = strings.Replace(result, "\n", "", -1)
	result = strings.Replace(result, "\\N", "\n", -1)
	//better := strings.Replace(re.FindString(result), "\n ", "", -1)
	//better = strings.Replace(better, "\\n", " ", -1)
	//better = strings.Replace(better, "\\", "", -1)

	//return trimField(better, "DESCRIPTION:")
	//return trimField(result, "DESCRIPTION:")
	return trimField(strings.Replace(result, "\r\n ", "", -1), "DESCRIPTION:")
}

func parseEventLocation(eventData *string) string {
	re, _ := regexp.Compile(`LOCATION:.*?\n`)
	result := re.FindString(*eventData)
	return trimField(result, "LOCATION:")
}

func parseEventRRule(eventData *string) string {
	re, _ := regexp.Compile(`RRULE:.*?\n`)
	result := re.FindString(*eventData)
	return trimField(result, "RRULE:")
}

func parseICalTimezone(eventData *string) time.Location {
	re, _ := regexp.Compile(`X-WR-TIMEZONE:.*?\n`)
	result := re.FindString(*eventData)

	// parse the timezone result to time.Location
	timezone := trimField(result, "X-WR-TIMEZONE:")
	// create location instance
	loc, err := time.LoadLocation(timezone)

	// if fails with the timezone => go Local
	if err != nil {
		loc, _ = time.LoadLocation("UTC")
	}
	return *loc
}

func parseMain(eventData *string, elementsP *[]Event, freq, href, color string) {
	eventStart, tzId := parseEventStart(eventData)
	eventEnd, tzId := parseEventEnd(eventData)
	start, _ := time.Parse(IcsFormat, startDate)
	end, _ := time.Parse(IcsFormat, endDate)
	//fmt.Println(start)

	var years, days, months int
	switch freq {
	case "DAILY":
		days = 1
		months = 0
		years = 0
		break
	case "WEEKLY":
		days = 7
		months = 0
		years = 0
		break
	case "MONTHLY":
		days = 0
		months = 1
		years = 0
		break
	case "YEARLY":
		days = 0
		months = 0
		years = 1
		break
	}
	//fmt.Println(eventStart)

	for {
		if inTimeSpan(start, end, eventStart) {
			data := Event{
				Href:        href,
				Color:       color,
				Start:       eventStart,
				End:         eventEnd,
				TZID:        tzId,
				Summary:     parseEventSummary(eventData),
				Description: parseEventDescription(eventData),
				Location:    parseEventLocation(eventData),
			}
			*elementsP = append(*elementsP, data)

		}

		if freq == "" {
			break
		}

		eventStart = eventStart.AddDate(years, months, days)
		eventEnd = eventEnd.AddDate(years, months, days)

		// TODO: support UNTIL
		if eventStart.After(end) {
			break
		}
	}
}
