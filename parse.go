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
	var datetime time.Time
	var tzID string

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
		if !strings.Contains(dt, "Z") {
			dt = fmt.Sprintf("%sZ", dt)

		}
		//fmt.Println(dt)
		location, _ := time.LoadLocation("Europe/Berlin")
		datetime, _ = time.Parse(IcsFormat, dt)

		// TODO: Solution for appointments with timezones. This here is bad.
		if tzID != "" {
			//loc, _ := time.LoadLocation(tzID)
			//datetime = datetime.UTC().In(loc)
			t = datetime
		} else {
			t = datetime.In(location)
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
	better := strings.Replace(re.FindString(*eventData), "\n ", "", -1)
	better = strings.Replace(better, "\\n", " ", -1)
	better = strings.Replace(better, "\\", "", -1)
	return trimField(better, "DESCRIPTION:")
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

func parseMain(eventData *string, elementsP *[]Event, freq, href string) {
	eventStart, tzId := parseEventStart(eventData)
	eventEnd, _ := parseEventEnd(eventData)
	start, _ := time.Parse(IcsFormatWholeDay, startDate)
	end, _ := time.Parse(IcsFormatWholeDay, endDate)

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

	for {
		if inTimeSpan(start, end, eventStart) {
			data := Event{
				Href:        href,
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
