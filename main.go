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
	"strings"
	"time"
)

const (
	ConfigDir  = ".config/qcal"
	CacheDir   = ".cache/qcal"
	dateFormat = "02.01.06"
	timeFormat = "15:04"
	RFC822     = "02.01.06 15:04"
	ColWhite   = "\033[1;37m"
	ColDefault = "\033[0m"
	ColGreen   = "\033[0;32m"
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

func showAppointments(startDate, endDate string) {
	config := getConf()
	/*xmlBody := `<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
		<d:prop>
			<d:getetag />
			<c:calendar-data />
		</d:prop>
		<c:filter>
			<c:comp-filter name="VCALENDAR">
				<c:comp-filter name="VEVENT">
					<c:time-range start="20191214T0815Z" end="20191216T0815Z"/>
				</c:comp-filter>
			</c:comp-filter>
		</c:filter>
	    </c:calendar-query>`*/
	//startDate := "20191214"
	//endDate := "20191216"
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

	//fmt.Println(string(xmlContent))
	err = xml.Unmarshal(xmlContent, &cald)
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < len(cald.Caldata); i++ {
		//fmt.Println(cald.Caldata[i].Data)
		elem := ParseICS(cald.Caldata[i].Data)
		fancyOutput(elem)
	}
}

func fancyOutput(elem *event) {
	//date := elem.Start.Format(dateFormat)
	starttime := elem.DTStart.Format(RFC822)
	fmt.Println(ColWhite + starttime + ColDefault + ` - ` + elem.Summary + ` (until ` + elem.DTEnd.Format(timeFormat) + `)`)
	// 		fmt.Println(elem.Summary)
	if elem.Description != "" {
		fmt.Println("Beschreibung: " + elem.Description)
	}
	if elem.Location != "" {
		fmt.Println("Ort: " + elem.Location)
	}
	fmt.Println()
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

	showAppointments(startDate, endDate)
	//	fmt.Printf("current time is :%s\n", curTime)
	//	fmt.Printf("calculated time is :%s", in10Days)
	//	fmt.Printf("calculated time is :%s", in10DaysFormat)

}
