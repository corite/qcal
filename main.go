package main

import (
	// 	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	// 	"regexp"
	"flag"
	"sort"
	"strings"
	"time"
)

const (
	ConfigDir      = ".config/qcal"
	CacheDir       = ".cache/qcal"
	dateFormat     = "02.01.06"
	dayMonthFormat = "02.01"
	timeFormat     = "15:04"
	RFC822         = "02.01.06 15:04"
	// ics date format ( describes a whole day)
	IcsFormat           = "20060102T150405Z"
	IcsFormatWholeDay   = "20060102"
	IcsFormatWholeMonth = "200601"
	IcsFormatMonthDay   = "0102"
	IcsFormatTime       = "T150405Z"
	Weekday             = "Mon"
	IcsFormatYear       = "2006"
	ColWhite            = "\033[1;37m"
	ColDefault          = "\033[0m"
	ColGreen            = "\033[0;32m"
)

type Caldata struct {
	XMLName xml.Name     `xml:"multistatus"`
	Caldata []Calelement `xml:"response"`
}

type Calelement struct {
	XMLName xml.Name `xml:"response"`
	Href    string   `xml:"href"`
	ETag    string   `xml:"propstat>prop>getetag"`
	Data    string   `xml:"propstat>prop>calendar-data"`
}

var err string
var homedir string = os.Getenv("HOME")
var configLocation string = (homedir + "/" + ConfigDir + "/config.json")
var cacheLocation string = (homedir + "/" + CacheDir)
var versionLocation string = (cacheLocation + "/version.json")

func fetchCalData(startDate, endDate string) Caldata {
	config := getConf()

	xmlBody := `<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
			<d:prop>
				<c:calendar-data />
			</d:prop>
			<c:filter>
				<c:comp-filter name="VCALENDAR"> 
					<c:comp-filter name="VEVENT">
						<c:time-range start="` + startDate + `T0000Z" end="` + endDate + `T2359Z"/>
					</c:comp-filter>
				</c:comp-filter>
			</c:filter>
		    </c:calendar-query>`

	req, err := http.NewRequest("REPORT", config.Url, strings.NewReader(xmlBody))
	req.SetBasicAuth(config.Username, config.Password)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	xmlContent, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	cald := Caldata{}
	// 	fmt.Println(string(xmlContent))
	err = xml.Unmarshal(xmlContent, &cald)
	if err != nil {
		log.Fatal(err)
	}

	return cald
}

func processEventsWeekly(eventData *string, startDate, endDate string, elementsP *[]Event) {
	eventStartDate, _ := parseEventStart(eventData)
	eventWeekday := eventStartDate.Format(Weekday)
	// weekday loop for recurring events
	var t time.Time
	t, _ = time.Parse(IcsFormatWholeDay, startDate)

	endDateFormated, _ := time.Parse(IcsFormatWholeDay, endDate)

	//for t; t.Equal(endDateFormated); t.AddDate(0, 0, 1) {
	for {
		if eventWeekday == t.Format(Weekday) {
			// copy event
			// add time from event origin to this date
			t2, _ := time.Parse(RFC822, t.Format(dateFormat)+` `+eventStartDate.Format(timeFormat))
			tend, _ := parseEventEnd(eventData)
			data := Event{
				Start: t2,
				End:   tend,
				//Freq:        parseEventRRule(cald.Caldata[i].Data, startDate, endDate),
				Summary:     parseEventSummary(eventData),
				Description: parseEventDescription(eventData),
				Location:    parseEventLocation(eventData),
			}
			//data.Href = *eventData.Href
			// put all in slice
			*elementsP = append(*elementsP, data)
		}
		// increment date
		t = t.AddDate(0, 0, 1)
		// end loop if incremented date = end date
		if t.Equal(endDateFormated) {
			break
		}

	}

}

func processEventsYearly(eventData *string, startDate, endDate string, elementsP *[]Event) {
	eventStartDate, _ := parseEventStart(eventData)
	startDateFormated, _ := time.Parse(IcsFormatWholeDay, startDate)
	endDateFormated, _ := time.Parse(IcsFormatWholeDay, endDate)

	for {
		// add date and time from event origin to this year
		t2, _ := time.Parse(IcsFormat, startDateFormated.Format(IcsFormatYear)+eventStartDate.Format(IcsFormatMonthDay)+eventStartDate.Format(IcsFormatTime))
		tend, _ := parseEventEnd(eventData)
		data := Event{
			Start: t2,
			End:   tend,
			//Freq:        parseEventRRule(cald.Caldata[i].Data, startDate, endDate),
			Summary:     parseEventSummary(eventData),
			Description: parseEventDescription(eventData),
			Location:    parseEventLocation(eventData),
		}
		//data.Href = cald.Caldata[i].Href
		// put all in slice
		*elementsP = append(*elementsP, data)
		// increment year
		startDateFormated = startDateFormated.AddDate(1, 0, 0)
		//fmt.Println(startDateFormated.Format(IcsFormat))
		// end loop if incremented date after end date
		if startDateFormated.After(endDateFormated) {
			break
		}
	}
}

func showAppointments(startDate, endDate string) {
	//elements := []*Event{}
	var elements []Event

	cald := fetchCalData(startDate, endDate)

	for i := 0; i < len(cald.Caldata); i++ {
		eventData := cald.Caldata[i].Data

		// week frequency
		result := eventFreqWeeklyRegex.FindString(eventData)
		if result != "" {
			processEventsWeekly(&eventData, startDate, endDate, &elements)
			continue
		}

		// year frequency
		result = eventFreqYearlyRegex.FindString(eventData)
		if result != "" {
			processEventsYearly(&eventData, startDate, endDate, &elements)
			continue
		}

		if result == "" {
			tstart, tz := parseEventStart(&eventData)
			tend, _ := parseEventEnd(&eventData)
			fmt.Println(eventData)
			fmt.Println(cald.Caldata[i].Href)

			data := Event{
				Start:       tstart,
				End:         tend,
				TZID:        tz,
				Summary:     parseEventSummary(&eventData),
				Description: parseEventDescription(&eventData),
				Location:    parseEventLocation(&eventData),
			}
			data.Href = cald.Caldata[i].Href
			// put all in slice
			elements = append(elements, data)
		}
	}

	// time.Time sort by start time for events
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Start.Before(elements[j].Start)
	})

	// pretty print
	for _, e := range elements {
		e.fancyOutput()
	}
}

func main() {
	var today string
	var startDate string
	var endDate string
	curTime := time.Now()
	todayFormat := curTime.Format(IcsFormatWholeDay)
	in10Days := curTime.Add(time.Hour * 240)
	in10DaysFormat := in10Days.Format(IcsFormatWholeDay)

	flag.StringVar(&today, "t", todayFormat, "Show appointments for today")
	flag.StringVar(&startDate, "start", todayFormat, "start date")
	flag.StringVar(&endDate, "end", in10DaysFormat, "end date")
	flag.Parse()

	//startDate = "20190801"
	//endDate = "20210902"
	showAppointments(startDate, endDate)
	//	fmt.Printf("current time is :%s\n", curTime)
}
