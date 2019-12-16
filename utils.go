package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
)

func TrimField(field, cutset string) string {
	re, _ := regexp.Compile(cutset)
	cutsetRem := re.ReplaceAllString(field, "")
	return strings.TrimRight(cutsetRem, "\r\n")
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
		writeLocalVersion()
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
	local := readLocalVersion()

	if p.CTag != local.CTag {
		//fmt.Println(p.CTag)
		return true
	}

	return false
}
