package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

type version struct {
	DisplayName  string
	CTag         string
	LastModified string
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

func WriteLocalVersion() bool {
	p := getProp()

	data := version{
		DisplayName:  p.DisplayName,
		CTag:         p.CTag,
		LastModified: p.LastModified,
	}

	file, _ := json.MarshalIndent(data, "", " ")

	// create cache dir if not exists
	os.MkdirAll(cacheLocation, os.ModePerm)

	err := ioutil.WriteFile(versionLocation, file, 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

func ReadLocalVersion() *version {
	if _, err := os.Stat(versionLocation); os.IsNotExist(err) {
		// if not yet exists get new version
		fmt.Println("No local version found. Getting remote...")
		// TODO get ics

		//downloadInitial()
		WriteLocalVersion()
	}

	versionData, err := ioutil.ReadFile(versionLocation)
	if err != nil {
		log.Fatal(err)
	}
	ver := version{}
	err = json.Unmarshal(versionData, &ver)

	return &ver
}

func NeedsUpdate() bool {
	p := getProp()
	local := ReadLocalVersion()

	if p.CTag != local.CTag {
		//fmt.Println(p.CTag)
		return true
	}

	return false
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
	version := ReadLocalVersion()

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
