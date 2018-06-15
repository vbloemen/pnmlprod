package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
)

type XES struct {
	XMLName xml.Name `xml:"log"`
	XTraces []XTrace `xml:"trace"`
}

type XTrace struct {
	XMLName xml.Name  `xml:"trace"`
	Key     XKeyValue `xml:"string"`
	Events  []XEvent  `xml:"event"`
}

type XKeyValue struct {
	XMLName xml.Name `xml:"string"`
	Key     string   `xml:"key,attr"`
	Value   string   `xml:"value,attr"`
}

type XEvent struct {
	XMLName   xml.Name    `xml:"event"`
	EventKeys []XKeyValue `xml:"string"`
}

func readXES(logfn string) []byte {
	file, err := os.Open(logfn)
	CheckError(err)
	defer file.Close()
	strFile := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		strFile += fmt.Sprintln(scanner.Text())
	}
	return []byte(strFile)
}

func ParseXES(logfn string) [][]string {
	contents := readPNML(logfn)
	var xes XES
	xml.Unmarshal(contents, &xes) // fill in XES contents

	var ret [][]string
	for _, t := range xes.XTraces {
		var trace []string
		for _, e := range t.Events {
			for _, kv := range e.EventKeys {
				if kv.Key == "concept:name" {
					trace = append(trace, kv.Value)
					break
				}
			}
		}
		ret = append(ret, trace)
	}
	return ret
}
