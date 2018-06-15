package main

import (
	"bufio"
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

const (
	LOG    string = "LOG"
	MODEL  string = "MODEL"
	SYNC   string = "SYNC"
	TAU    string = "TAU" // don't use "tau" as ltsmin sees this as invisible
	TAUSYM string = "τ"
)

var (
	MOVES []string = []string{LOG, MODEL, SYNC, TAU}
)

type PNML struct {
	XMLName xml.Name `xml:"pnml"`
	Net     Net      `xml:"net"`
}

type Net struct {
	XMLName      xml.Name `xml:"net"`
	ID           string   `xml:"id,attr"`
	Type         string   `xml:"type,attr"`
	Name         string   `xml:"name>text"`
	Page         Page     `xml:"page"`
	FinalMarking Marking  `xml:"finalmarkings>marking"`
}

type Marking struct {
	XMLName xml.Name `xml:"marking"`
	MPlaces []MPlace `xml:"place"`
}

type MPlace struct {
	XMLName    xml.Name `xml:"place"`
	ID         string   `xml:"idref,attr"`
	TokenCount string   `xml:"text"`
}

type Page struct {
	XMLName     xml.Name     `xml:"page"`
	ID          string       `xml:"id,attr"`
	Places      []Place      `xml:"place"`
	Transitions []Transition `xml:"transition"`
	Arcs        []Arc        `xml:"arc"`
}

type Place struct {
	XMLName        xml.Name `xml:"place"`
	ID             string   `xml:"id,attr"`
	Name           string   `xml:"name>text"`
	InitialMarking string   `xml:"initialMarking>text"`
	FinalMarking   string   `xml:"finalMarking>text"`
	Type           string   `xml:"type>text"` // added for {model,log} places
}

type Transition struct {
	XMLName  xml.Name `xml:"transition"`
	ID       string   `xml:"id,attr"`
	Name     string   `xml:"name>text"`
	OrigName string   `xml:"origname>text"`
	Type     string   `xml:"type>text"`     // added for {model,log,sync,tau}-moves
	Selected string   `xml:"selected>text"` // for DOT printing
}

type Arc struct {
	XMLName xml.Name `xml:"arc"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name>text"`
	Source  string   `xml:"source,attr"`
	Target  string   `xml:"target,attr"`
}

// Parsing

func readPNML(filename string) []byte {
	file, err := os.Open(filename)
	CheckError(err)
	defer file.Close()
	// In case the encoding ISO-8859-1 is used, just change it to UTF-8
	// the XML parser doesn't handle the ISO encoding very well..
	xml := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"
	strFile := ""
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "<?xml") {
			strFile += fmt.Sprintln(xml)
		} else {
			strFile += fmt.Sprintln(scanner.Text())
		}
	}
	return []byte(strFile)
}

// I/O

func (pn *PNML) Print() {
	fmt.Println(pn.Net.Name)
	for i, _ := range pn.Net.Page.Places {
		fmt.Println(pn.Net.Page.Places[i])
	}
	for _, trans := range pn.Net.Page.Transitions {
		fmt.Println(trans)
	}
	for _, arc := range pn.Net.Page.Arcs {
		fmt.Println(arc)
	}
	fmt.Println("Final marking")
	for _, mplace := range pn.Net.FinalMarking.MPlaces {
		fmt.Println(mplace)
	}
}

func dotTypeColor(t, selected string) string {
	if selected == "true" {
		return "firebrick1"
	}
	switch t {
	case LOG:
		return "darkgoldenrod1"
	case MODEL:
		return "lightskyblue"
	case SYNC:
		return "chartreuse"
	case TAU:
		return "grey"
	default:
		return "white"
	}
}

func (pn *PNML) dotArcColor(arc *Arc) string {
	// search for the transition
	for _, trans := range pn.Net.Page.Transitions {
		if arc.Source == trans.ID || arc.Target == trans.ID {
			if trans.Selected == "true" {
				return "firebrick"
			}
			switch trans.Type {
			case LOG:
				return "darkorange"
			case MODEL:
				return "blue"
			case SYNC:
				return "forestgreen"
			case TAU:
				return "grey27"
			default:
				return "black"
			}
		}
	}
	return "black"
}

func (pn *PNML) PrintDOT(filename string) {
	ret := "digraph g {\n"
	ret += "  rankdir=\"LR\";\n" // horizontal layout
	// NB: add initial/final marking info?
	// draw log nodes and trans
	ret += "  subgraph cluster_l {\n    style=invisible\n"
	for _, place := range pn.Net.Page.Places {
		if place.Type == LOG {
			ret += fmt.Sprintf("    %s [label=\"%s\", shape=circle"+
				", style=\"filled,solid\", fillcolor=\"%s\""+
				", fontname=\"Courier-Bold\"];\n",
				place.ID, place.ID, dotTypeColor(place.Type, ""))
		}
	}
	for _, trans := range pn.Net.Page.Transitions {
		if trans.Type == LOG {
			ret += fmt.Sprintf("    %s [label=\"%s\", shape=box"+
				", style=\"filled,solid\", fillcolor=\"%s\""+
				", fontname=\"Courier-Bold\"];\n",
				trans.ID, trans.OrigName, dotTypeColor(trans.Type,
					trans.Selected))
		}
	}
	ret += "  }\n"
	// draw model nodes and trans
	ret += "  subgraph cluster_m {\n    style=invisible\n"
	for _, place := range pn.Net.Page.Places {
		if place.Type == MODEL {
			ret += fmt.Sprintf("    %s [label=\"%s\", shape=circle"+
				", style=\"filled,solid\", fillcolor=\"%s\""+
				", fontname=\"Courier-Bold\"];\n",
				place.ID, place.ID, dotTypeColor(place.Type, ""))
		}
	}
	for _, trans := range pn.Net.Page.Transitions {
		if trans.Type == MODEL {
			ret += fmt.Sprintf("    %s [label=\"%s\", shape=box"+
				", style=\"filled,solid\", fillcolor=\"%s\""+
				", fontname=\"Courier-Bold\"];\n",
				trans.ID, trans.OrigName, dotTypeColor(trans.Type,
					trans.Selected))
		} else if trans.Type == TAU {
			ret += fmt.Sprintf("    %s [label=\"%s\", shape=box"+
				", style=\"filled,solid\", fillcolor=\"%s\""+
				", fontname=\"Courier-Bold\"];\n",
				trans.ID, TAUSYM, dotTypeColor(trans.Type, trans.Selected))
		}
	}
	ret += "  }\n"
	// draw sync trans
	for _, trans := range pn.Net.Page.Transitions {
		if trans.Type == SYNC {
			ret += fmt.Sprintf("  %s [label=\"%s\", shape=box"+
				", style=\"filled,solid\", fillcolor=\"%s\""+
				", fontname=\"Courier-Bold\"];\n",
				trans.ID, trans.OrigName, dotTypeColor(trans.Type,
					trans.Selected))
		}
	}
	for _, arc := range pn.Net.Page.Arcs {
		ret += fmt.Sprintf("  %s -> %s [penwidth=2, color=\"%s\""+
			", fontcolor=\"black\"];\n",
			arc.Source, arc.Target, pn.dotArcColor(&arc))
	}

	ret += "}\n"
	WriteFile(filename, ret)
}
