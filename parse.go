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
	IcsFormat = "20060102T150405Z"
	// Y-m-d H:i:S time format
	YmdHis = "2006-01-02 15:04:05"
	// ics date format ( describes a whole day)
	IcsFormatWholeDay = "20060102"
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
	//re, _ := regexp.Compile(fmt.Sprintf(`%s;TZID=(.*?)?(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))

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

		if result == nil || len(result) < 4 {
			return t, tzID
		}

		tzID = result[2]
		dt := result[4]
		if !strings.Contains(dt, "Z") {
			dt = fmt.Sprintf("%sZ", dt)
		}
		t, _ = time.Parse(IcsFormat, dt)
	}

	return t, tzID
}

// parses the event start time
func parseEventStart(eventData string) (time.Time, string) {
	return parseTimeField("DTSTART", eventData)
}

// parses the event end time
func parseEventEnd(eventData string) (time.Time, string) {
	return parseTimeField("DTEND", eventData)
}

func parseEventRRule(eventData string) string {
	// 	freq := trimField(eventFreqWeeklyRegex.FindString(eventData), "RRULE:")
	result := eventFreqWeeklyRegex.FindString(eventData)
	if result != "" {
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
func parseEventSummary(eventData string) string {
	re, _ := regexp.Compile(`SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, `SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?:`)
}

func parseEventDescription(eventData string) string {
	re, _ := regexp.Compile(`DESCRIPTION:.*?\n(?:\s+.*?\n)*`)
	better := strings.Replace(re.FindString(eventData), "\n ", "", -1)
	better = strings.Replace(better, "\\n", " ", -1)
	better = strings.Replace(better, "\\", "", -1)
	return trimField(better, "DESCRIPTION:")
}

func parseEventLocation(eventData string) string {
	re, _ := regexp.Compile(`LOCATION:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "LOCATION:")
}

func ParseICS(icsElem string) *Event {
	// 	// starttime
	// 	var dtstart time.Time
	//
	// 	var start string
	// 	re, _ := regexp.Compile(`DTSTART;TZID=(.*?)?(;VALUE=DATE-TIME)?:(.*?)\n`)
	// 	if re.FindString(icsElem) == "" {
	// 		// 		fmt.Println("----------jetzt---")
	// 		re, _ = regexp.Compile(`DTSTART?:(.*?)\n`)
	// 		start = trimField(re.FindString(icsElem), `DTSTART?:`)
	// 	} else {
	// 		start = trimField(re.FindString(icsElem), `DTSTART;TZID=(.*?)?(;VALUE=DATE-TIME)?:`)
	// 	}
	// 	if !strings.Contains(start, "Z") {
	// 		start = fmt.Sprintf("%sZ", start)
	// 	}
	// 	dtstart, _ = time.Parse(IcsFormat, start)
	//
	// 	// endtime
	// 	var dtend time.Time
	// 	var end string
	// 	re, _ = regexp.Compile(`DTEND;TZID=(.*?)?(;VALUE=DATE-TIME)?:(.*?)\n`)
	// 	if re.FindString(icsElem) == "" {
	// 		// 		fmt.Println("----------jetzt---")
	// 		re, _ = regexp.Compile(`DTEND?:(.*?)\n`)
	// 		end = trimField(re.FindString(icsElem), `DTEND?:`)
	// 	} else {
	// 		end = trimField(re.FindString(icsElem), `DTEND;TZID=(.*?)?(;VALUE=DATE-TIME)?:`)
	// 	}
	// 	if !strings.Contains(end, "Z") {
	// 		end = fmt.Sprintf("%sZ", end)
	// 	}
	// 	dtend, _ = time.Parse(IcsFormat, end)

	tstart, tz := parseEventStart(icsElem)
	tend, _ := parseEventEnd(icsElem)

	//	fmt.Println(parseEventRRule(icsElem))

	data := Event{
		Start:       tstart,
		End:         tend,
		TZID:        tz,
		Freq:        parseEventRRule(icsElem),
		Summary:     parseEventSummary(icsElem),
		Description: parseEventDescription(icsElem),
		Location:    parseEventLocation(icsElem),
	}

	fmt.Println(data.Freq)
	return &data
}
