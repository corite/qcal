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
	WholeDay    bool
}

func trimField(field, cutset string) string {
	re, _ := regexp.Compile(cutset)
	cutsetRem := re.ReplaceAllString(field, "")
	return strings.TrimRight(cutsetRem, "\r\n")
}

// parses the event start time
func parseTimeField(fieldName string, eventData string) (time.Time, string) {
	reWholeDay, _ := regexp.Compile(fmt.Sprintf(`%s;VALUE=DATE:.*?\n`, fieldName))
	//re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))?(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))
	re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))

	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time
	var tzID string

	if resultWholeDay != "" {
		// whole day event
		modified := trimField(resultWholeDay, fmt.Sprintf("%s;VALUE=DATE:", fieldName))
		t, _ = time.Parse(IcsFormatWholeDay, modified)
	} else {
		// event that has start hour and minute
		result := re.FindStringSubmatch(eventData)

		// why? TODO
		if result == nil || len(result) < 4 {
			re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))?(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))
			result = re.FindStringSubmatch(eventData)
		}

		tzID = result[2]
		dt := result[4]
		if !strings.Contains(dt, "Z") {
			dt = fmt.Sprintf("%sZ", dt)

		}
		//fmt.Println(dt)
		t, _ = time.Parse(IcsFormat, dt)
	}

	return t, tzID
}

// parses the event start time
func parseEventStart(eventData *string) (time.Time, string) {
	return parseTimeField("DTSTART", *eventData)
}

// parses the event end time
func parseEventEnd(eventData *string) (time.Time, string) {
	return parseTimeField("DTEND", *eventData)
}

func parseEventRRule(eventData string, startDate string, endDate string) string {
	// 	freq := trimField(eventFreqWeeklyRegex.FindString(eventData), "RRULE:")
	result := eventFreqWeeklyRegex.FindString(eventData)
	if result != "" {
		eventStartDate, _ := parseEventStart(&eventData)
		eventWeekday := eventStartDate.Format(Weekday)
		// weekday loop for recurring events
		var t time.Time
		t, _ = time.Parse(IcsFormatWholeDay, startDate)
		var endDateFormated time.Time
		endDateFormated, _ = time.Parse(IcsFormatWholeDay, endDate)

		//for t; t.Before(endDateFormated); t.AddDate(0, 0, 1) {
		for {
			if eventWeekday == t.Format(Weekday) {
				// copy event
				//fmt.Println(t.Format(Weekday))
			}
			// increment date
			t = t.AddDate(0, 0, 1)
			// end for loop if incremented date = end date
			if t.Equal(endDateFormated) {
				break
			}

		}
		// 		fmt.Println(t.Format(Weekday))
		fmt.Println(endDate)

		// 		curWeekday := startDate.Format(Weekday)

		fmt.Println(eventWeekday)
		// 		fmt.Println(t)
		return trimField(result, "RRULE:FREQ=")
	}
	result = eventFreqYearlyRegex.FindString(eventData)
	if result != "" {
		return trimField(result, "RRULE:FREQ=")
	}
	//fmt.Println(result)

	//fmt.Println(freq)
	return trimField(result, "RRULE:FREQ=")
}

// parses the event summary
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

func ParseICS(icsElem string, startDate string, endDate string) *Event {
	tstart, tz := parseEventStart(&icsElem)
	tend, _ := parseEventEnd(&icsElem)

	//	fmt.Println(parseEventRRule(icsElem))

	data := Event{
		Start:       tstart,
		End:         tend,
		TZID:        tz,
		Freq:        parseEventRRule(icsElem, startDate, endDate),
		Summary:     parseEventSummary(&icsElem),
		Description: parseEventDescription(&icsElem),
		Location:    parseEventLocation(&icsElem),
	}

	fmt.Println(data.Freq)
	return &data
}
