package main

import (
	//"encoding/json"
	"fmt"
	// 	"log"
	"regexp"
	"strings"
	"time"

	duration "github.com/channelmeter/iso8601duration"
)

var (
	eventRRuleRegex      = regexp.MustCompile(`RRULE:.*?\n`)
	freqRegex            = regexp.MustCompile(`FREQ=.*?;`)
	eventSummaryRegex    = regexp.MustCompile(`SUMMARY:.*?\n`)
	eventFreqWeeklyRegex = regexp.MustCompile(`RRULE:FREQ=WEEKLY\n`)
	eventFreqYearlyRegex = regexp.MustCompile(`RRULE:FREQ=YEARLY\n`)
)

// unixtimestamp
const (
	uts = "1136239445"
	//ics date time format
	// Y-m-d H:i:S time format
	YmdHis = "2006-01-02 15:04:05"
)

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
func splitIcal(ical string) []string {
	splits := regexp.MustCompile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	//reEvent, _ := regexp.Compile(`(BEGIN:VEVENT(.*\n)*?END:VEVENT\r?\n)`)
	Events := splits.FindAllString(ical, -1)
	/*for i := range Events {
		//fmt.Println(eventData)
		fmt.Println(i)
		fmt.Println(Events[i])
	}*/
	//fmt.Println(Events[1])
	//os.Exit(1)
	return Events
}

func parseIcalName(eventData string) string {
	re, _ := regexp.Compile(`X-WR-CALNAME:.*?\n`)
	result := re.FindString(eventData)
	return trimField(result, "X-WR-CALNAME:")
}

func parseTimeField(fieldName string, eventData string) (time.Time, string) {
	reWholeDay, _ := regexp.Compile(fmt.Sprintf(`%s;VALUE=DATE:.*?\n`, fieldName))
	//re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))?(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))
	// correct regex: .+:(.+)$
	re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.+))?(;VALUE=DATE-TIME)?:(.+?)\n`, fieldName))
	//re, _ := regexp.Compile(fmt.Sprintf(`%s(;TZID=(.*?))(;VALUE=DATE-TIME)?:(.*?)\n`, fieldName))

	resultWholeDay := reWholeDay.FindString(eventData)
	var t time.Time
	var thisTime time.Time
	//var thisTime time.Time
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
		//fmt.Println(result)

		if result == nil || len(result) < 4 {
			return t, tzID
		}

		tzID = result[2]
		//fmt.Println(tzID)
		dt := strings.Trim(result[4], "\r") // trim these newlines!

		if strings.HasSuffix(dt, "Z") {
			// If string ends in 'Z', timezone is UTC
			format = "20060102T150405Z"
			thisTime, _ := time.Parse(format, dt)
			//fmt.Println(thisTime)
			t = thisTime.Local()
		} else if tzID != "" {
			format = "20060102T150405"
			location, err := time.LoadLocation(tzID)
			//fmt.Println(location)
			// if tzID not readable use configured timezone
			if err != nil {
				location, _ = time.LoadLocation(config.Timezone)
				// timezone from defines gives CEST, which is not working with parseinlocation:
				//location, _ = time.LoadLocation(timezone)
			}
			// set foreign timezone
			thisTime, _ = time.ParseInLocation(format, dt, location)
			// convert to local timezone
			//t = time.In(myLocation)
			t = thisTime.Local()
		} else {
			// Else, consider the timezone is local the parser
			format = "20060102T150405"
			t, _ = time.Parse(format, dt)
			//fmt.Println(t)
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

func parseEventDuration(eventData *string) time.Duration {
	reDuration, _ := regexp.Compile(`DURATION:.*?\n`)
	result := reDuration.FindString(*eventData)
	trimmed := trimField(result, "DURATION:")
	parsedDuration, err := duration.FromString(trimmed)
	var output time.Duration

	if err == nil {
		output = parsedDuration.ToDuration()
	}

	return output
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
	result = strings.Replace(result, "\n ", "", -1)

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

func parseEventAttendees(eventData *string) []string {
	//re, _ := regexp.Compile(`ATTENDEE;.*?\n`)
	re, _ := regexp.Compile(`ATTENDEE;.+\"(.+?)\".*\n`)
	attendeesstring := re.FindAllString(*eventData, -1)
	var attendees []string

	for i := range attendeesstring {
		//fmt.Println(eventData)
		result := re.FindStringSubmatch(attendeesstring[i])
		attendees = append(attendees, result[1])

		//attendee := trimField(attendees[i], `ATTENDEE;.*\"`)
		//fmt.Println(result[1])
	}

	return attendees
}

func parseEventRRule(eventData *string) string {
	re, _ := regexp.Compile(`RRULE:.*?\n`)
	result := re.FindString(*eventData)
	return trimField(result, "RRULE:")
}

func parseEventFreq(eventData *string) string {
	re, _ := regexp.Compile(`FREQ=[^;]*(;){0,1}`)
	result := re.FindString(parseEventRRule(eventData))
	return trimField(result, `(FREQ=|;)`)
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

func parseMain(eventData *string, elementsP *[]Event, href, color string) {
	eventStart, tzId := parseEventStart(eventData)
	eventEnd, tzId := parseEventEnd(eventData)
	eventDuration := parseEventDuration(eventData)
	freq := parseEventFreq(eventData)

	if eventEnd.Before(eventStart) {
		eventEnd = eventStart.Add(eventDuration)
	}

	start, _ := time.Parse(IcsFormat, startDate)
	end, _ := time.Parse(IcsFormat, endDate)
	//fmt.Println(eventStart)

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
				Attendees:   parseEventAttendees(eventData),
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
