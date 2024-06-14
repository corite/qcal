package main

import (
	// 	"bytes"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

var config = getConf()

func fetchCalData(calNo int, wg *sync.WaitGroup) {
	xmlBody := `<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
			<d:prop><c:calendar-data /><d:getetag /></d:prop>
			<c:filter><c:comp-filter name="VCALENDAR"><c:comp-filter name="VEVENT">
				<c:time-range start="` + startDate + `Z" end="` + endDate + `Z" />
			 </c:comp-filter></c:comp-filter></c:filter>
		    </c:calendar-query>`

	//fmt.Println(xmlBody)
	reqType := "REPORT"

	if config.Calendars[calNo].Username == "" {
		reqType = "GET" // some servers don't like REPORT
		xmlBody = ""
	}

	req, _ := http.NewRequest(reqType, config.Calendars[calNo].Url, strings.NewReader(xmlBody))

	if config.Calendars[calNo].Username != "" {
		req.SetBasicAuth(config.Calendars[calNo].Username, config.Calendars[calNo].password())
		req.Header.Add("Depth", "1") // needed for SabreDAV
		req.Header.Add("Prefer", "return-minimal")
		req.Header.Add("Content-Type", "application/xml; charset=utf-8")
	}

	/*tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	cli := &http.Client{Transport: tr}*/

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	xmlContent, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	//fmt.Println(string(xmlContent))

	cald := Caldata{}
	err = xml.Unmarshal(xmlContent, &cald)
	if err != nil { // if no XML
		//log.Fatal(err)
		//fmt.Println("noxml")
		eventData := splitIcal(string(xmlContent))
		//fmt.Println(eventData)
		//log.Fatal(err)
		for i := range eventData {
			eventHref := ""
			eventColor := Colors[calNo]
			parseMain(&eventData[i], &elements, eventHref, eventColor)
		}
	} else {
		for i := range cald.Caldata {
			eventData := cald.Caldata[i].Data
			eventHref := cald.Caldata[i].Href
			eventColor := Colors[calNo]
			//fmt.Println(eventData)
			//fmt.Println(i)

			eventData, _ = explodeEvent(&eventData) // vevent only
			parseMain(&eventData, &elements, eventHref, eventColor)
		}
	}

	wg.Done()
}

func showAppointments(singleCal string) {
	var wg sync.WaitGroup         // use waitgroups to fetch calendars in parallel
	wg.Add(len(config.Calendars)) // waitgroup length = num calendars
	for i := range config.Calendars {
		if singleCal == fmt.Sprintf("%v", i) || singleCal == "all" { // sprintf because convert int to string
			go fetchCalData(i, &wg)
		} else {
			wg.Done()
		}
	}
	wg.Wait()

	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Start.Before(elements[j].Start) // time.Time sort by start time for events
	})

	if len(elements) == 0 {
		log.Fatal("no events") // get out if nothing found
	}

	for _, e := range elements {
		e.fancyOutput() // pretty print
	}
}

func createAppointment(calNumber string, appointmentData string, recurrence string) {
	curTime := time.Now()
	dataArr := strings.Split(appointmentData, " ")
	var startDate string
	var endDate string
	var startTime string
	var endTime string
	var summaryIdx int
	var isDayEvent bool
	var calRec string
	var dtStartString string
	var dtEndString string

	calNo, _ := strconv.ParseInt(calNumber, 0, 64)
	// no username, no write.
	if config.Calendars[calNo].Username == "" {
		log.Fatal("You can't write to iCal calendars")
	}

	// first block - start (and possible end-) date
	if (isNumeric(dataArr[0])) && (len(dataArr[0]) == 8) {
		startDate = dataArr[0]
	} else {
		log.Fatal("Wrong date/time syntax. Please check help.")
	}
	// second block - start time or end date
	if isNumeric(dataArr[1]) {
		// if is start time
		if len(dataArr[1]) == 4 {
			endDate = dataArr[0]
			startTime = dataArr[1]
			summaryIdx = 3
			// if second block is start time, third block needs to be end time
			if len(dataArr[2]) == 4 {
				endTime = dataArr[2]
			} else {
				log.Fatal("Wrong date/time syntax. Please check help.")
			}
			// if is end date
		} else if len(dataArr[1]) == 8 {
			endDate = dataArr[1]
			isDayEvent = true
			summaryIdx = 2
		}
	} else {
		// if only start date is given treat as one whole day appointment
		endDateObject, _ := time.Parse(IcsFormatWholeDay, startDate)
		endDate = endDateObject.AddDate(0, 0, 1).Format(IcsFormatWholeDay)
		isDayEvent = true
		summaryIdx = 1
	}
	//fmt.Println(summaryIdx)

	summary := dataArr[summaryIdx]
	for i := range dataArr {
		if i > summaryIdx {
			summary = summary + ` ` + dataArr[i]
		}
	}
	newElem := genUUID() + `.ics`
	//tzName, _ := time.Now().Zone()
	//fmt.Printf("name: [%v]\toffset: [%v]\n", tzName, tzOffset)
	tzName, e := time.LoadLocation(config.Timezone)
	checkError(e)

	timezoneString := fmt.Sprintf("%v", tzName)

	if isDayEvent {
		dtStartString = fmt.Sprintf("VALUE=DATE:%v", startDate)
		dtEndString = fmt.Sprintf("VALUE=DATE:%v", endDate)
	} else {
		dtStartString = fmt.Sprintf("TZID=%v:%vT%v00", tzName, startDate, startTime)
		dtEndString = fmt.Sprintf("TZID=%v:%vT%v00", tzName, endDate, endTime)
	}
	//check frequency
	switch recurrence {
	case "d":
		calRec = "\nRRULE:FREQ=DAILY"
	case "w":
		calRec = "\nRRULE:FREQ=WEEKLY"
	case "m":
		calRec = "\nRRULE:FREQ=MONTHLY"
	case "y":
		calRec = "\nRRULE:FREQ=YEARLY"
	default:
		calRec = ""
	}

	var calSkel = `BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//qcal
BEGIN:VTIMEZONE
TZID:` + timezoneString + `
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
DTSTART;` + dtStartString + `
DTEND;` + dtEndString + `
DTSTAMP:` + curTime.UTC().Format(IcsFormat) + `Z
SUMMARY:` + summary + calRec + `
END:VEVENT
END:VCALENDAR`
	//fmt.Println(calSkel)
	//os.Exit(3)

	req, _ := http.NewRequest("PUT", config.Calendars[calNo].Url+newElem, strings.NewReader(calSkel))
	req.SetBasicAuth(config.Calendars[calNo].Username, config.Calendars[calNo].password())
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
	//curTimeDay := curTime.Truncate(24 * time.Hour).Add(-1) // NO! UTC only
	// remove time, substract 1 msec for whole day appointments
	curTimeDay := time.Date(curTime.Year(), curTime.Month(), curTime.Day(), 0, 0, 0, 0, time.Local).Add(-1)
	//fmt.Println(curTimeDay)
	toFile := false

	flag.StringVar(&startDate, "s", curTimeDay.UTC().Format(IcsFormat), "Start date")                                  // default today
	flag.StringVar(&endDate, "e", curTimeDay.UTC().AddDate(0, 0, config.DefaultNumDays).Format(IcsFormat), "End date") // default 1 month
	flag.BoolVar(&showInfo, "i", false, "Show additional info like description and location for appointments")
	flag.BoolVar(&showFilename, "f", false, "Show appointment filename for editing or deletion")
	flag.BoolVar(&displayFlag, "p", false, "Print ICS file piped to qcal (for CLI mail tools like mutt)")
	flag.BoolVar(&showWeekday, "wd", false, "Show weekday")
	calNumber := flag.String("c", "all", "Show only single calendar (number)")
	showToday := flag.Bool("t", false, "Show appointments for today")
	show7days := flag.Bool("7", false, "Show 7 days from now")
	version := flag.Bool("v", false, "Show version")
	showMinutes := flag.Int("cron", 15, "Crontab mode. Show only appointments in the next n minutes.")
	recurrence := flag.String("r", "", "Recurrency for new appointments. Use d,w,m,y with \"-n\"")
	showCalendars := flag.Bool("l", false, "List configured calendars with their corresponding numbers (for \"-c\")")
	appointmentFile := flag.String("u", "", "Upload appointment file. Provide filename and use with \"-c\"")
	appointmentDelete := flag.String("delete", "", "Delete appointment. Get filename with \"-f\" and use with \"-c\"")
	appointmentDump := flag.String("d", "", "Dump raw appointment data. Get filename with \"-f\" and use with \"-c\"")
	appointmentEdit := flag.String("edit", "", "Edit + upload appointment data. Get filename with \"-f\" and use with \"-c\"")
	appointmentData := flag.String("n", "", "Add a new appointment. Check README.md for syntax")
	flag.Parse()
	flagset := make(map[string]bool) // map for flag.Visit. get bools to determine set flags
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if *showToday {
		endDate = curTimeDay.UTC().AddDate(0, 0, 1).Format(IcsFormat) // today till tomorrow
	}
	if *show7days {
		endDate = curTimeDay.UTC().AddDate(0, 0, 7).Format(IcsFormat) // today till 7 days
	}
	if *showCalendars {
	}
	if flagset["cron"] {
		startDate = curTime.UTC().Format(IcsFormat)
		endDate = curTime.UTC().Add(time.Minute * time.Duration(*showMinutes)).Format(IcsFormat)
		showColor = false
	}

	if flagset["l"] {
		getProp()
	} else if flagset["n"] {
		createAppointment(*calNumber, *appointmentData, *recurrence)
	} else if flagset["delete"] {
		deleteEvent(*calNumber, *appointmentDelete)
	} else if flagset["d"] {
		dumpEvent(*calNumber, *appointmentDump, toFile)
	} else if flagset["p"] {
		displayICS()
	} else if flagset["edit"] {
		editEvent(*calNumber, *appointmentEdit)
	} else if flagset["u"] {
		eventEdit := false
		uploadICS(*calNumber, *appointmentFile, eventEdit)
	} else if *version {
		fmt.Print("qcal ")
		fmt.Println(qcalversion)
	} else {
		//startDate = "20210301"
		//endDate = "20210402"
		showAppointments(*calNumber)
	}
}
