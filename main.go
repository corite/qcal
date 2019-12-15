// comment1
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
	Username  = "psic4t"
	Password  = "Ich hab Haus!"
	Url       = "https://mail.data.haus/dav/psic4t/63c75aea-d9f3-40c0-4617-a5a624a3ba64/"
	ConfigDir = ".config/qcal"
	CacheDir  = ".cache/qcal"
)

var err string

type config struct {
	Username string
	Password string
	Url      string
}

type props struct {
	XMLName      xml.Name `xml:"multistatus"`
	DisplayName  string   `xml:"response>propstat>prop>displayname"`
	CTag         string   `xml:"response>propstat>prop>getctag"`
	ETag         string   `xml:"response>propstat>prop>getetag"`
	LastModified string   `xml:"response>propstat>prop>getlastmodified"`
}

func getConf() *config {
	homedir := os.Getenv("HOME")
	configLocation := (homedir + "/" + ConfigDir + "/config.json")
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

func getProp() {
	req, errHttp := http.NewRequest("PROPFIND", Url, nil)
	req.SetBasicAuth(Username, Password)

	cli := &http.Client{}
	resp, errHttp := cli.Do(req)
	if errHttp != nil {
		log.Fatal(errHttp)
	}

	xmlContent, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	p := props{}

	errXml := xml.Unmarshal(xmlContent, &p)
	if errXml != nil {
		panic(errXml)
	}

	fmt.Printf("%#v", p)
	fmt.Println()
	fmt.Println(p.ETag)
}

func main() {
	config := getConf()
	fmt.Println(config.Username)
	fmt.Println(config.Url)

	//fmt.Println(&conf.Username)
	// 	getProp()
}
