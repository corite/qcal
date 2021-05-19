package main

import (
	"crypto/rand"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
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
		fmt.Print(ColYellow + e.Start.Format(dateFormat) + ColDefault + ` `)
		fmt.Printf(`%6s`, ` `)
		//fmt.Println(e)
		//if e.Start.Format(dateFormat) == e.End.Format(dateFormat) {
		if e.Start.Add(time.Hour*24) == e.End {
			fmt.Println(ColDefault + e.Summary + ColDefault)
		} else {
			fmt.Println(ColDefault + e.Summary + ColDefault + ` (until ` + e.End.Format(dateFormat) + `)`)
		}
	} else {
		fmt.Print(ColYellow + e.Start.Format(RFC822) + ColDefault + ` `)
		fmt.Println(ColDefault + e.Summary + ColDefault + ` (until ` + e.End.Format(timeFormat) + `)`)
	}

	if showInfo {
		if e.Description != "" {
			fmt.Printf(`%15s`, ` `)
			fmt.Println(`Beschreibung: ` + e.Description)
		}
		if e.Location != "" {
			fmt.Printf(`%15s`, ` `)
			fmt.Println("Ort: " + e.Location)
		}
	}
	if showFilename {
		if e.Href != "" {
			fmt.Println(path.Base(e.Href))
		}
	}
	//fmt.Println()
}

func genUUID() (uuid string) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return
	}
	uuid = fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])

	return
}

func strToInt(str string) (int, error) {
	nonFractionalPart := strings.Split(str, ".")
	return strconv.Atoi(nonFractionalPart[0])
}

func deleteEvent(calNumber string, eventFilename string) (status string) {
	config := getConf()

	calNo, _ := strconv.ParseInt(calNumber, 0, 64)
	//fmt.Println(config.Calendars[calNo].Url + eventFilename)

	req, _ := http.NewRequest("DELETE", config.Calendars[calNo].Url+eventFilename, nil)
	req.SetBasicAuth(config.Calendars[calNo].Username, config.Calendars[calNo].Password)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Status)

	return
}

func dumpEvent(calNumber string, eventFilename string, toFile bool) (status string) {
	config := getConf()

	calNo, _ := strconv.ParseInt(calNumber, 0, 64)
	//fmt.Println(config.Calendars[calNo].Url + eventFilename)

	req, _ := http.NewRequest("GET", config.Calendars[calNo].Url+eventFilename, nil)
	req.SetBasicAuth(config.Calendars[calNo].Username, config.Calendars[calNo].Password)

	cli := &http.Client{}
	resp, err := cli.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println(resp.Status)
	xmlContent, _ := ioutil.ReadAll(resp.Body)

	if toFile {
		// create cache dir if not exists
		os.MkdirAll(cacheLocation, os.ModePerm)
		err := ioutil.WriteFile(cacheLocation+"/"+eventFilename, xmlContent, 0644)
		if err != nil {
			log.Fatal(err)
		}
		return eventFilename + " written"
	} else {
		fmt.Println(string(xmlContent))
		return
	}
}

func uploadICS(calNumber string, eventFilename string) (status string) {
	config := getConf()

	calNo, _ := strconv.ParseInt(calNumber, 0, 64)
	//fmt.Println(config.Calendars[calNo].Url + eventFilename)

	eventICS, err := ioutil.ReadFile(cacheLocation + "/" + eventFilename)
	if err != nil {
		log.Fatal(err)
	}

	req, _ := http.NewRequest("PUT", config.Calendars[calNo].Url+eventFilename, strings.NewReader(string(eventICS)))
	req.SetBasicAuth(config.Calendars[calNo].Username, config.Calendars[calNo].Password)
	req.Header.Add("Content-Type", "text/calendar; charset=utf-8")

	cli := &http.Client{}
	resp, err := cli.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Status)

	return
}
