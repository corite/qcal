package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func getConf() *config {
	configData, err := ioutil.ReadFile(configLocation)
	if err != nil {
		log.Fatal(err)
	}

	conf := config{}
	err = json.Unmarshal(configData, &conf)
	//fmt.Println(conf)
	if err != nil {
		fmt.Println("error:", err)
	}

	return &conf
}

func getProp() props {
	// TODO
	config := getConf()
	p := props{}
	for i := range config.Calendars {
		//req, err := http.NewRequest("REPORT", config.Url, strings.NewReader(xmlBody))
		req, err := http.NewRequest("PROPFIND", config.Calendars[i].Url, nil)
		req.SetBasicAuth(config.Calendars[i].Username, config.Calendars[i].Password)

		cli := &http.Client{}
		resp, err := cli.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		xmlContent, _ := ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()

		//fmt.Println(string(xmlContent))
		err = xml.Unmarshal(xmlContent, &p)
		if err != nil {
			panic(err)
		}

		//fmt.Printf(xml.Unmarshal(xmlContent, &p))
		fmt.Println(`[` + fmt.Sprintf("%v", i) + `] - ` + p.DisplayName)
		fmt.Println(p.Color)
	}

	return p
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
			fmt.Println(ColDefault + e.Summary + ColDefault)
		} else {
			fmt.Println(ColDefault + e.Summary + ColDefault + ` (until ` + e.End.Format(dateFormat) + `)`)
		}
	} else {
		fmt.Print(ColGreen + e.Start.Format(RFC822) + ColDefault + ` `)
		fmt.Println(ColDefault + e.Summary + ColDefault + ` (until ` + e.End.Format(timeFormat) + `)`)
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
