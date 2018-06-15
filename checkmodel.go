package main

import (
	"encoding/xml"
	"fmt"
)

func CheckModel(filename string) {
	modelcontents := readPNML(filename)
	var pn PNML
	xml.Unmarshal(modelcontents, &pn) // fill in PNML contents

	places := len(pn.Net.Page.Places)
	trans := len(pn.Net.Page.Transitions)
	arcs := len(pn.Net.Page.Arcs)

	fmt.Printf("%d,%d,%d\n", places, trans, arcs)
	_ = places
	_ = trans
	_ = arcs
}
