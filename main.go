package main

import (
	// 	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

func fetchCalData(Url, Username, Password string, cald *Caldata, wg *sync.WaitGroup) {
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

	req, err := http.NewRequest("REPORT", Url, strings.NewReader(xmlBody))
	req.SetBasicAuth(Username, Password)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	xmlContent, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	//fmt.Println(string(xmlContent))
	err = xml.Unmarshal(xmlContent, &cald)
	if err != nil {
		log.Fatal(err)
	}
	wg.Done()
}

func showAppointments(singleCal string) {
	config := getConf()
	cald := Caldata{}
	var elements []Event

	// use waitgroups to fetch calendars in parallel
	var wg sync.WaitGroup
	wg.Add(len(config.Calendars)) // waitgroup length = num calendars
	for i := range config.Calendars {
		if singleCal == fmt.Sprintf("%v", i) || singleCal == "all" { // sprintf bc convert int to string
			//fmt.Println("Fetching...")
			go fetchCalData(config.Calendars[i].Url, config.Calendars[i].Username, config.Calendars[i].Password, &cald, &wg)
		} else {
			wg.Done()
		}
	}
	wg.Wait()

	//for i := 0; i < len(cald.Caldata); i++ {
	for i := range cald.Caldata {
		eventData := cald.Caldata[i].Data
		eventHref := cald.Caldata[i].Href
		// fmt.Println(eventHref)

		eventData, _ = explodeEvent(&eventData) // vevent only

		reFr, _ := regexp.Compile(`FREQ=[^;]*(;){0,1}`)
		freq := trimField(reFr.FindString(parseEventRRule(&eventData)), `(FREQ=|;)`)

		parseMain(&eventData, &elements, freq, eventHref)
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

func createAppointment(calNumber string, appointmentData string) {
	config := getConf()
	curTime := time.Now()
	dataArr := strings.Split(appointmentData, " ")
	startDate := dataArr[0]
	endDate := dataArr[0]
	startTime := dataArr[1]
	endTime := dataArr[2]
	summary := dataArr[3]
	for i := range dataArr {
		if i > 3 {
			summary = summary + ` ` + dataArr[i]
		}
	}
	newElem := genUUID() + `.ics`
	calNo, _ := strconv.ParseInt(calNumber, 0, 64)
	dtStart, _ := time.Parse(IcsFormat, startDate+`T`+startTime+`00Z`)
	dtEnd, _ := time.Parse(IcsFormat, endDate+`T`+endTime+`00Z`)
	//DTSTART;TZID=` + timezone + `:` + startDate + `T` + startTime + `00Z
	//DTSTART;TZID=` + timezone + `:` + startDate + `T` + startTime + `00
	//DTEND;TZID=` + timezone + `:` + endDate + `T` + endTime + `00

	var calSkel = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//qcal
METHOD:PUBLISH
BEGIN:VTIMEZONE
TZID:Europe/Berlin
BEGIN:STANDARD
DTSTART:16011028T030000
RRULE:FREQ=YEARLY;BYDAY=-1SU;BYMONTH=10
TZOFFSETFROM:+0200
TZOFFSETTO:+0100
END:STANDARD
BEGIN:DAYLIGHT
DTSTART:16010325T020000
RRULE:FREQ=YEARLY;BYDAY=-1SU;BYMONTH=3
TZOFFSETFROM:+0100
TZOFFSETTO:+0200
END:DAYLIGHT
END:VTIMEZONE
BEGIN:VEVENT
UID:` + curTime.UTC().Format(IcsFormat) + `-` + newElem + `
DTSTART;TZID=Europe/Berlin:` + dtStart.UTC().Format(IcsFormat) + ` 
DTEND;TZID=Europe/Berlin:` + dtEnd.UTC().Format(IcsFormat) + `
DTSTAMP:` + curTime.UTC().Format(IcsFormat) + `
SUMMARY:` + summary + `
END:VEVENT
END:VCALENDAR`
	//fmt.Println(calSkel)

	req, _ := http.NewRequest("PUT", config.Calendars[calNo].Url+newElem, strings.NewReader(calSkel))
	req.SetBasicAuth(config.Calendars[calNo].Username, config.Calendars[calNo].Password)
	req.Header.Add("Content-Type", "text/calendar; charset=utf-8")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Status)

}

func main() {
	curTime := time.Now()

	flag.StringVar(&startDate, "s", curTime.Format(IcsFormatWholeDay), "start date")              // default today
	flag.StringVar(&endDate, "e", curTime.AddDate(0, 2, 0).Format(IcsFormatWholeDay), "end date") // default 2 month
	flag.BoolVar(&showInfo, "i", false, "Show additional info like summary or location for appointments")
	flag.BoolVar(&showFilename, "f", false, "Show appointment filename for editing or deletion")
	calNumber := flag.String("c", "all", "Show only single calendar (number)")
	showToday := flag.Bool("t", false, "Show appointments for today")
	show7days := flag.Bool("7", false, "Show 7 days from now")
	showCalendars := flag.Bool("C", false, "Show available calendars")
	appointmentFile := flag.String("d", "", "Delete appointment. Get filename with \"-f\" and use with -c")
	appointmentDump := flag.String("dump", "", "Dump raw  appointment data. Get filename with \"-f\" and use with -c")
	appointmentData := flag.String("n", "20210425 0800 0900 bla blubb foo bar", "Add a new appointment. Syntax: yyyymmdd hhmm hhmm subject")
	flag.Parse()
	flagset := make(map[string]bool) // map for flag.Visit. get bools to determine set flags
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *showToday {
		endDate = curTime.AddDate(0, 0, 1).Format(IcsFormatWholeDay) // today till tomorrow
	}
	if *show7days {
		endDate = curTime.AddDate(0, 0, 7).Format(IcsFormatWholeDay) // today till 7 days
	}
	if *showCalendars {
	}

	if flagset["C"] {
		getProp()
	} else if flagset["n"] {
		createAppointment(*calNumber, *appointmentData)
	} else if flagset["d"] {
		deleteEvent(*calNumber, *appointmentFile)
	} else if flagset["dump"] {
		dumpEvent(*calNumber, *appointmentDump)
	} else {
		//startDate = "20210301"
		//endDate = "20210402"
		//createAppointment(*appointmentData)
		showAppointments(*calNumber)
		//	fmt.Printf("current time is :%s\n", curTime)
	}
	/*	switch flagset {
		case flagset["n"]:
			createAppointment(*appointmentData)
		case flagset["C"]:
			getProp()
		default:
			//startDate = "20210301"
			//endDate = "20210402"
			showAppointments(singleCal)
			//      fmt.Printf("current time is :%s\n", curTime)
		}
	*/
}
