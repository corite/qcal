package main

import (
	"encoding/xml"
	"os"
	"time"
)

var err string
var homedir string = os.Getenv("HOME")
var editor string = os.Getenv("EDITOR")
var configLocation string = (homedir + "/" + ConfigDir + "/config.json")
var cacheLocation string = (homedir + "/" + CacheDir)
var versionLocation string = (cacheLocation + "/version.json")
var timezone, _ = time.Now().Zone()
var xmlContent []byte
var showInfo bool
var showFilename bool
var startDate string
var endDate string
var summary string
var toFile bool

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
	ColYellow           = "\033[1;33m"
)

/*
type config struct {
	Username string
	Password string
	Url      string
}
*/

type config struct {
	Calendars []struct {
		Username string
		Password string
		Url      string
	}
	Timezone string
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
