package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

type config struct {
	Username string
	Password string
	Url      string
}

/*
type config struct {
	Calendars struct {
		Username string
		Password string
		Url      string
	} `json:"Calendars"`
}
*/
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

	conf := config{}
	err = json.Unmarshal(configData, &conf)
	if err != nil {
		fmt.Println("error:", err)
	}

	return &conf
}

func inTimeSpan(start, end, check time.Time) bool {
	return check.After(start) && check.Before(end)
}

//func fancyOutput(elem *event) {
func (e Event) fancyOutput() {
	// whole day or greater
	if e.Start.Format(timeFormat) == e.End.Format(timeFormat) {
		fmt.Print(ColGreen + e.Start.Format(dateFormat) + ColDefault + ` `)
		fmt.Printf(`%6s`, ` `)
		//fmt.Println(e)
		//if e.Start.Format(dateFormat) == e.End.Format(dateFormat) {
		if e.Start.Add(time.Hour*24) == e.End {
			fmt.Println(ColWhite + e.Summary + ColDefault)
		} else {
			fmt.Println(ColWhite + e.Summary + ColDefault + ` (until ` + e.End.Format(dateFormat) + `)`)
		}
	} else {
		fmt.Print(ColGreen + e.Start.Format(RFC822) + ColDefault + ` `)
		fmt.Println(ColWhite + e.Summary + ColDefault + ` (until ` + e.End.Format(timeFormat) + `)`)
	}

	//fmt.Println(e.Href)
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
