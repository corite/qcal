package main

import (
	// 	"bytes"
	"encoding/xml"
	"flag"
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
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
	ColBlue             = "\033[1;34m"
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
var configLocation string = (homedir + "/" + ConfigDir + "/config-2.json")
var cacheLocation string = (homedir + "/" + CacheDir)
var versionLocation string = (cacheLocation + "/version.json")
var timezone, _ = time.Now().Zone()

var calSkel = `BEGIN:VCALENDAR
		VERSION:2.0
		CALSCALE:GREGORIAN
		PRODID:-//qcal
		BEGIN:VEVENT
		TZID:` + timezone + `
		DTSTART;TZID=` + timezone + `:20191011T193000Z
		DTEND;TZID=` + timezone + `:20191011T123000Z
		DTSTAMP:20190930T141136Z
		SUMMARY:Training mit Eric
		END:VEVENT
		END:VCALENDAR`

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

	cald := Caldata{}

	for i := 0; i < len(config.Calendars); i++ {
		//calendars := []config.Calendars{}
		//for i := range calendars {
		//req, err := http.NewRequest("REPORT", config.Url, strings.NewReader(xmlBody))
		req, err := http.NewRequest("REPORT", config.Calendars[i].Url, strings.NewReader(xmlBody))
		req.SetBasicAuth(config.Calendars[i].Username, config.Calendars[i].Password)

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
	}

	return cald
}

func showAppointments(startDate, endDate string) {
	var elements []Event

	cald := fetchCalData(startDate, endDate)

	for i := 0; i < len(cald.Caldata); i++ {
		eventData := cald.Caldata[i].Data
		eventHref := cald.Caldata[i].Href

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
	var today string
	var startDate string
	var endDate string
	curTime := time.Now()
	todayFormat := curTime.Format(IcsFormatWholeDay)
	//in10Days := curTime.AddDate(0, 0, 10)
	//in10DaysFormat := in10Days.Format(IcsFormatWholeDay)
	in2Months := curTime.AddDate(0, 2, 0)
	in2MonthsFormat := in2Months.Format(IcsFormatWholeDay)

	flag.StringVar(&today, "t", todayFormat, "Show appointments for today")
	flag.StringVar(&startDate, "start", todayFormat, "start date")
	//flag.StringVar(&endDate, "end", in10DaysFormat, "end date")
	flag.StringVar(&endDate, "end", in2MonthsFormat, "end date")
	flag.Parse()

	//startDate = "20210301"
	//endDate = "20210402"
	//getConf()
	showAppointments(startDate, endDate)
	//	fmt.Printf("current time is :%s\n", curTime)
}
