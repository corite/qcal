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
	"strings"
	"sync"
	"time"
)

func fetchCalData(startDate, endDate, Url, Username, Password string, cald *Caldata, wg *sync.WaitGroup) {
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

// use waitgroups to fetch calendars in parallel
func fetchCalDataParallel(startDate, endDate, singleCal string) Caldata {
	config := getConf()
	cald := Caldata{}

	var wg sync.WaitGroup
	wg.Add(len(config.Calendars)) // waitgroup length = num calendars
	for i := range config.Calendars {
		if singleCal == fmt.Sprintf("%v", i) || singleCal == "all" { // sprintf bc convert int to string
			//fmt.Println("Fetching...")
			go fetchCalData(startDate, endDate, config.Calendars[i].Url, config.Calendars[i].Username, config.Calendars[i].Password, &cald, &wg)
		} else {
			wg.Done()
		}
	}
	wg.Wait()
	//fmt.Println(&cald.Caldata)
	return cald
}

func showAppointments(startDate, endDate, singleCal string) {
	var elements []Event

	cald := fetchCalDataParallel(startDate, endDate, singleCal)

	for i := 0; i < len(cald.Caldata); i++ {
		eventData := cald.Caldata[i].Data
		eventHref := cald.Caldata[i].Href
		//fmt.Println(eventHref)

		eventData, _ = explodeEvent(&eventData) // vevent only

		reFr, _ := regexp.Compile(`FREQ=[^;]*(;){0,1}`)
		freq := trimField(reFr.FindString(parseEventRRule(&eventData)), `(FREQ=|;)`)

		parseMain(&eventData, &elements, startDate, endDate, freq, eventHref)
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
	var startDate string
	var endDate string
	var singleCal string
	curTime := time.Now()

	flag.StringVar(&singleCal, "c", "all", "Show only single calendar (number)")
	showToday := flag.Bool("t", false, "Show appointments for today")
	show7days := flag.Bool("7", false, "Show 7 days from now")
	flag.BoolVar(&showInfo, "i", false, "Show additional info like summary or location for appointments")
	showCalendars := flag.Bool("C", false, "Show available calendars")
	flag.StringVar(&startDate, "s", curTime.Format(IcsFormatWholeDay), "start date")              // default today
	flag.StringVar(&endDate, "e", curTime.AddDate(0, 2, 0).Format(IcsFormatWholeDay), "end date") // default 2 month
	flag.Parse()

	if *showToday {
		endDate = curTime.AddDate(0, 0, 1).Format(IcsFormatWholeDay) // today till tomorrow
	}
	if *show7days {
		endDate = curTime.AddDate(0, 0, 7).Format(IcsFormatWholeDay) // today till 7 days
	}
	if *showCalendars {
		getProp()
	} else {

		//startDate = "20210301"
		//endDate = "20210402"
		showAppointments(startDate, endDate, singleCal)
		//	fmt.Printf("current time is :%s\n", curTime)
	}
}
