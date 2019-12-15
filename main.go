package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
)

const (
	ConfigDir = ".config/qcal"
	CacheDir  = ".cache/qcal"
)

var err string
var homedir string = os.Getenv("HOME")
var versionLocation string = (homedir + "/" + CacheDir + "/version.json")
var configLocation string = (homedir + "/" + ConfigDir + "/config.json")

type config struct {
	Username string
	Password string
	Url      string
}

type version struct {
	CTag         string
	LastModified string
}

type props struct {
	XMLName      xml.Name `xml:"multistatus"`
	DisplayName  string   `xml:"response>propstat>prop>displayname"`
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

	// 	fmt.Printf("%#v", p)
	// 	fmt.Println(p.ETag)
	return &p
}

func writeLocalVersion() bool {
	p := getProp()

	data := version{
		CTag:         p.CTag,
		LastModified: p.LastModified,
	}

	file, _ := json.MarshalIndent(data, "", " ")

	err := ioutil.WriteFile(versionLocation, file, 0644)
	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}

func readLocalVersion() string {
	if _, err := os.Stat(versionLocation); os.IsNotExist(err) {
		// if not yet exists get new version
		fmt.Println("No local version found. Getting...")
		writeLocalVersion()
	}

	versionData, err := ioutil.ReadFile(versionLocation)
	if err != nil {
		log.Fatal(err)
	}
	ver := version{}
	err = json.Unmarshal(versionData, &ver)

	return ver.CTag
}

func compareVersion(ctag string) bool {
	// check if version file exists
	if _, err := os.Stat(versionLocation); os.IsNotExist(err) {
		return false
	}

	// read version file
	versionData, err := ioutil.ReadFile(versionLocation)
	if err != nil {
		log.Fatal(err)
		return false
	}
	ver := version{}
	err = json.Unmarshal(versionData, &ver)

	return true
}

func main() {
	//fmt.Println(&conf.Username)

	//p := getProp()
	r := readLocalVersion()
	fmt.Println(r)
}
