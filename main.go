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
	"regexp"
	"strings"
)

const (
	ConfigDir = ".config/qcal"
	CacheDir  = ".cache/qcal"
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

type version struct {
	DisplayName  string
	CTag         string
	LastModified string
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

func DownloadInitial() {
	config := getConf()
	xmlBody := `<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
			<d:prop>
				<d:getetag />
				<c:calendar-data />
			</d:prop>
			<c:filter>
				<c:comp-filter name="VCALENDAR" /> 
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

	err = xml.Unmarshal(xmlContent, &cald)
	if err != nil {
		panic(err)
	}

	// read version file for cal name
	version := NeadLocalVersion()

	// create dir if not exist for cal
	os.MkdirAll(cacheLocation+"/"+version.DisplayName, os.ModePerm)

	fmt.Println("Downloading all calendar entries...")

	for i := 0; i < len(cald.Caldata); i++ {
		// split href
		s := strings.Split(cald.Caldata[i].Href, "/")
		filename := s[len(s)-1]

		// string to byte for writing
		data := []byte(cald.Caldata[i].Data)
		err := ioutil.WriteFile(cacheLocation+"/"+version.DisplayName+"/"+filename, data, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}
	fmt.Println("Downloading done.")
}

func showAppointments() {
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
	startDate := "20191214"
	endDate := "20191216"
	xmlBody := `<c:calendar-query xmlns:d="DAV:" xmlns:c="urn:ietf:params:xml:ns:caldav">
			<d:prop>
				<d:getetag />
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

	// read version file for cal name
	for i := 0; i < len(cald.Caldata); i++ {
		//fmt.Println(cald.Caldata[i].Data)
		re, _ := regexp.Compile(`SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?.*?\n`)
		summary := re.FindString(cald.Caldata[i].Data)

		re, _ = regexp.Compile(`DESCRIPTION:.*?\n(?:\s+.*?\n)*`)
		description := re.FindString(cald.Caldata[i].Data)

		fmt.Println(TrimField(summary, `SUMMARY(?:;LANGUAGE=[a-zA-Z\-]+)?:`) + description)
	}
}

func main() {
	//fmt.Println(&conf.Username)

	//p := getProp()
	//n := needsUpdate()
	// 	r := readLocalVersion()
	//fmt.Println(p.Href)
	//downloadInitial()
	showAppointments()
}
