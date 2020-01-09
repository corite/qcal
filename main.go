package main

import (
	// 	"bytes"
	"encoding/json"
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

var err string
var homedir string = os.Getenv("HOME")
var configLocation string = (homedir + "/" + ConfigDir + "/config.json")
var cacheLocation string = (homedir + "/" + CacheDir)
var versionLocation string = (cacheLocation + "/version.json")

type config struct {
	Username string
	Password string
	Url      string
}

type props struct {
	XMLName      xml.Name `xml:"multistatus"`
	Href         string   `xml:"response>href"`
	DisplayName  string   `xml:"response>propstat>prop>displayname"`
	Color        string   `xml:"response>propstat>prop>calendar-color"`
	CTag         string   `xml:"response>propstat>prop>getctag"`
	ETag         string   `xml:"response>propstat>prop>getetag"`
	LastModified string   `xml:"response>propstat>prop>getlastmodified"`
}

func getConf() *config {
	configData, err := ioutil.ReadFile(configLocation)
	if err != nil {
		log.Fatal(err)
	}

	//var conf config
	conf := config{}
	err = json.Unmarshal(configData, &conf)
	if err != nil {
		fmt.Println("error:", err)
	}

	//fmt.Println(conf.Username)
	return &conf
}

func getProp() *props {
	config := getConf()
	req, err := http.NewRequest("PROPFIND", config.Url, nil)
	req.SetBasicAuth(config.Username, config.Password)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	xmlContent, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	p := props{}

	err = xml.Unmarshal(xmlContent, &p)
	if err != nil {
		panic(err)
	}

	//fmt.Printf("%#v", string(xmlContent))
	return &p
}

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

func showAppointments(startDate, endDate string) {
	elements := []*Event{}

	cald := fetchCalData(startDate, endDate)

	for i := 0; i < len(cald.Caldata); i++ {
		// week frequency
		result := eventFreqWeeklyRegex.FindString(cald.Caldata[i].Data)
		if result != "" {
			eventStartDate, _ := parseEventStart(cald.Caldata[i].Data)
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
					tend, _ := parseEventEnd(cald.Caldata[i].Data)
					data := Event{
						Start: t2,
						End:   tend,
						//Freq:        parseEventRRule(cald.Caldata[i].Data, startDate, endDate),
						Summary:     parseEventSummary(cald.Caldata[i].Data),
						Description: parseEventDescription(cald.Caldata[i].Data),
						Location:    parseEventLocation(cald.Caldata[i].Data),
					}
					data.Href = cald.Caldata[i].Href
					// put all in slice
					elements = append(elements, &data)
				}
				// increment date
				t = t.AddDate(0, 0, 1)
				// end loop if incremented date = end date
				if t.Equal(endDateFormated) {
					break
				}

			}
		}

		// year frequency
		result = eventFreqYearlyRegex.FindString(cald.Caldata[i].Data)
		if result != "" {
			eventStartDate, _ := parseEventStart(cald.Caldata[i].Data)
			startDateFormated, _ := time.Parse(IcsFormatWholeDay, startDate)
			endDateFormated, _ := time.Parse(IcsFormatWholeDay, endDate)

			for {
				// copy event
				// add date and time from event origin to this year
				t2, _ := time.Parse(IcsFormat, startDateFormated.Format(IcsFormatYear)+eventStartDate.Format(IcsFormatMonthDay)+eventStartDate.Format(IcsFormatTime))
				tend, _ := parseEventEnd(cald.Caldata[i].Data)
				data := Event{
					Start: t2,
					End:   tend,
					//Freq:        parseEventRRule(cald.Caldata[i].Data, startDate, endDate),
					Summary:     parseEventSummary(cald.Caldata[i].Data),
					Description: parseEventDescription(cald.Caldata[i].Data),
					Location:    parseEventLocation(cald.Caldata[i].Data),
				}
				data.Href = cald.Caldata[i].Href
				// put all in slice
				elements = append(elements, &data)
				// increment year
				startDateFormated = startDateFormated.AddDate(1, 0, 0)
				//fmt.Println(startDateFormated.Format(IcsFormat))
				// end loop if incremented date after end date
				if startDateFormated.After(endDateFormated) {
					break
				}

			}
		}

		if result == "" {
			tstart, tz := parseEventStart(cald.Caldata[i].Data)
			tend, _ := parseEventEnd(cald.Caldata[i].Data)

			data := Event{
				Start: tstart,
				End:   tend,
				TZID:  tz,
				//Freq:        parseEventRRule(cald.Caldata[i].Data, startDate, endDate),
				Summary:     parseEventSummary(cald.Caldata[i].Data),
				Description: parseEventDescription(cald.Caldata[i].Data),
				Location:    parseEventLocation(cald.Caldata[i].Data),
			}
			data.Href = cald.Caldata[i].Href
			// put all in slice
			elements = append(elements, &data)
		}
	}

	// time.Time sort by start time for events
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Start.Before(elements[j].Start)
	})

	// pretty print
	for _, e := range elements {
		//fancyOutput(value)
		e.fancyOutput()
	}

}

//func fancyOutput(elem *event) {
func (e Event) fancyOutput() {
	// whole day or greater
	if e.Start.Format(timeFormat) == e.End.Format(timeFormat) {
		fmt.Print(ColGreen + e.Start.Format(dateFormat) + ColDefault + ` `)
		fmt.Printf(`%6s`, ` `)
		fmt.Println(e.Summary)
		/*		if e.Start.Format(dateFormat) == e.End.Format(dateFormat) {
					fmt.Println(e.Summary)
				} else {
					fmt.Println(e.Summary + ` (until ` + e.End.Format(dateFormat) + `)`)
				}*/
	} else {
		fmt.Print(ColGreen + e.Start.Format(RFC822) + ColDefault + ` `)
		fmt.Println(e.Summary + ` (until ` + e.End.Format(timeFormat) + `)`)
	}

	// 		fmt.Println(elem.Summary)
	if e.Description != "" {
		fmt.Printf(`%15s`, ` `)
		fmt.Println(`Beschreibung: ` + e.Description)
	}
	if e.Location != "" {
		fmt.Printf(`%15s`, ` `)
		fmt.Println("Ort: " + e.Location)
	}
	//fmt.Println()
}

func main() {
	//p := getProp()

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

	//startDate = "20191001"
	//endDate = "20210902"
	showAppointments(startDate, endDate)
	//	fmt.Printf("current time is :%s\n", curTime)
	//	fmt.Printf("calculated time is :%s", in10Days)
	//	fmt.Printf("calculated time is :%s", in10DaysFormat)

}
