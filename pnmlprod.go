package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"strings"
)

func readLog(logfn string) [][]string {
	var ret [][]string
	if strings.HasSuffix(logfn, ".csv") {
		file, err := os.Open(logfn)
		CheckError(err)
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			split := strings.Split(scanner.Text(), ",")
			if len(split) != 0 && split[0] != "" {
				ret = append(ret, split)
			}
		}
	} else if strings.HasSuffix(logfn, ".xes") {
		ret = ParseXES(logfn)
	} else {
		CheckError(errors.New("Unknown file extention for log file"))
	}
	//fmt.Printf("There are %d log traces\nSizes:", len(ret))
	//for _, tr := range ret {
	//	fmt.Printf(" %d", len(tr))
	//}
	//fmt.Printf("\n")
	return ret
}

func CreatePNMLProduct(modelfn, logfn, outdir string) {
	_, err := os.Stat(outdir)
	if os.IsNotExist(err) {
		CheckError(errors.New("Directory doesn't exist '" + outdir + "'"))
	}
	modelcontents := readPNML(modelfn)
	logtraces := readLog(logfn)
	for i, logtrace := range logtraces {
		// creating the synchronous product
		var pn PNML
		xml.Unmarshal(modelcontents, &pn) // fill in PNML contents
		pn.PostProcessPNML()              // fill in initial markings etc.
		pn.PrintDOT(fmt.Sprintf("%s.dot", modelfn[:len(modelfn)-5]))
		pn.AddLog(logtrace)
		pn.PostProcessProduct()
		// I/O
		//fmt.Println("Printing synchronous PN")

		//mg := pn.CreateMarkingGraph()
		//mg.PrintDOT(fmt.Sprintf("%s-mg.dot", modelfn[:len(modelfn)-5]))

		//pn.Print()
		pn.PrintDOT(fmt.Sprintf(outdir+"/syncmodel-%d.dot", i))
		output, err := xml.Marshal(pn) // output PNML contents
		CheckError(err)
		WriteFile(fmt.Sprintf(outdir+"/syncmodel-%d.pnml", i), string(output))
		WriteFile(fmt.Sprintf(outdir+"/invariant-%d.txt", i),
			pn.GenerateInvariant())
	}
}

// should be called after unmarshalling
func (pn *PNML) PostProcessPNML() {
	// change initialMarkings from "" to "0"
	for i, _ := range pn.Net.Page.Places {
		if len(pn.Net.Page.Places[i].InitialMarking) == 0 {
			pn.Net.Page.Places[i].InitialMarking = "0"
		}
		pn.Net.Page.Places[i].Type = MODEL
	}
	// change tau transitions to be TAU and add type TAU
	// add type MODEL to all non-tau transitions
	for i, _ := range pn.Net.Page.Transitions {
		if strings.Contains(pn.Net.Page.Transitions[i].Name, "tau") {
			pn.Net.Page.Transitions[i].Name = TAUSYM // τ
			pn.Net.Page.Transitions[i].Type = TAU
		} else {
			pn.Net.Page.Transitions[i].Type = MODEL
		}
		pn.Net.Page.Transitions[i].OrigName = pn.Net.Page.Transitions[i].Name
	}
	if len(pn.Net.FinalMarking.MPlaces) == 0 {
		CheckError(errors.New("Unable to parse final markings." +
			" Is the PNML an accepting net?"))
	}
}

func (pn *PNML) PostProcessProduct() {
	// set type as name
	for i, _ := range pn.Net.Page.Transitions {
		pn.Net.Page.Transitions[i].OrigName = pn.Net.Page.Transitions[i].Name
		pn.Net.Page.Transitions[i].Name = pn.Net.Page.Transitions[i].Type
	}
	// add final marking to the places
	for _, m := range pn.Net.FinalMarking.MPlaces {
		for i, p := range pn.Net.Page.Places {
			if m.ID == p.ID {
				pn.Net.Page.Places[i].FinalMarking = m.TokenCount
			}
		}
	}
}

type transArcs struct {
	ID   string
	Name string
	In   []string // Place ID
	Out  []string // Place ID
}

// returns a slice of matching model transitions
func (pn *PNML) matchingModelTrans(name string) []transArcs {
	// NB: only search model trans
	ret := []transArcs{}
	for _, trans := range pn.Net.Page.Transitions {
		if trans.Type == MODEL && trans.Name == name {
			// search for in and out arcs
			inArcs := []string{}
			outArcs := []string{}
			for _, arc := range pn.Net.Page.Arcs {
				if arc.Target == trans.ID {
					inArcs = append(inArcs, arc.Source)
				}
				if arc.Source == trans.ID {
					outArcs = append(outArcs, arc.Target)
				}
			}

			ret = append(ret, transArcs{ID: trans.ID, Name: name,
				In: inArcs, Out: outArcs})
		}
	}
	return ret
}

// assumes the log trace is given as a CSV: "a,b,c,tau,s"
func (pn *PNML) AddLog(logtrace []string) {
	if len(logtrace) == 0 {
		return
	}
	p := &Place{XMLName: xml.Name{Space: "", Local: "place"},
		ID: "logp0", Name: "logp0", InitialMarking: "1", Type: LOG}
	pn.Net.Page.Places = append(pn.Net.Page.Places, *p)
	for logid, letter := range logtrace {
		// add log moves
		tl := &Transition{XMLName: xml.Name{Space: "", Local: "transition"},
			ID: fmt.Sprintf("logt%d", logid), Name: letter, Type: LOG}
		p := &Place{XMLName: xml.Name{Space: "", Local: "place"},
			ID:             fmt.Sprintf("logp%d", logid+1),
			Name:           fmt.Sprintf("logp%d", logid+1),
			InitialMarking: "0", Type: LOG}
		a1 := &Arc{XMLName: xml.Name{Space: "", Local: "arc"},
			ID:     fmt.Sprintf("arcp%d", logid),
			Name:   fmt.Sprintf("arcp%d", logid),
			Source: fmt.Sprintf("logp%d", logid),
			Target: fmt.Sprintf("logt%d", logid)}
		a2 := &Arc{XMLName: xml.Name{Space: "", Local: "arc"},
			ID:     fmt.Sprintf("arct%d", logid),
			Name:   fmt.Sprintf("arct%d", logid),
			Source: fmt.Sprintf("logt%d", logid),
			Target: fmt.Sprintf("logp%d", logid+1)}
		pn.Net.Page.Transitions = append(pn.Net.Page.Transitions, *tl)
		pn.Net.Page.Places = append(pn.Net.Page.Places, *p)
		pn.Net.Page.Arcs = append(pn.Net.Page.Arcs, *a1, *a2)

		// add sync moves for ALL matches in the model
		for taid, ta := range pn.matchingModelTrans(letter) {

			ts := &Transition{XMLName: xml.Name{Space: "",
				Local: "transition"}, ID: fmt.Sprintf("logs%dn%d", logid, taid),
				Name: letter, Type: SYNC}
			a3 := &Arc{XMLName: xml.Name{Space: "", Local: "arc"},
				ID:     fmt.Sprintf("arcp%dn%d", logid, taid),
				Name:   fmt.Sprintf("arcp%dn%d", logid, taid),
				Source: fmt.Sprintf("logp%d", logid),
				Target: fmt.Sprintf("logs%dn%d", logid, taid)}
			a4 := &Arc{XMLName: xml.Name{Space: "", Local: "arc"},
				ID:     fmt.Sprintf("arct%dn%d", logid, taid),
				Name:   fmt.Sprintf("arct%dn%d", logid, taid),
				Source: fmt.Sprintf("logs%dn%d", logid, taid),
				Target: fmt.Sprintf("logp%d", logid+1)}

			for inid, in := range ta.In {
				ai := &Arc{XMLName: xml.Name{Space: "", Local: "arc"},
					ID:     fmt.Sprintf("arcin%dn%dn%d", logid, taid, inid),
					Name:   fmt.Sprintf("arcin%dn%dn%d", logid, taid, inid),
					Source: in,
					Target: fmt.Sprintf("logs%dn%d", logid, taid)}
				pn.Net.Page.Arcs = append(pn.Net.Page.Arcs, *ai)
			}
			for outid, out := range ta.Out {
				ao := &Arc{XMLName: xml.Name{Space: "", Local: "arc"},
					ID:     fmt.Sprintf("arcout%dn%dn%d", logid, taid, outid),
					Name:   fmt.Sprintf("arcout%dn%dn%d", logid, taid, outid),
					Source: fmt.Sprintf("logs%dn%d", logid, taid),
					Target: out}
				pn.Net.Page.Arcs = append(pn.Net.Page.Arcs, *ao)
			}

			pn.Net.Page.Transitions = append(pn.Net.Page.Transitions, *ts)
			pn.Net.Page.Arcs = append(pn.Net.Page.Arcs, *a3, *a4)
		}

		// update final marking
		mp := &MPlace{XMLName: xml.Name{Space: "", Local: "place"},
			ID: fmt.Sprintf("logp%d", logid), TokenCount: "0"}
		pn.Net.FinalMarking.MPlaces = append(pn.Net.FinalMarking.MPlaces, *mp)
	}
	// final place of log is in final marking
	mp := &MPlace{XMLName: xml.Name{Space: "", Local: "place"},
		ID: fmt.Sprintf("logp%d", len(logtrace)), TokenCount: "1"}
	pn.Net.FinalMarking.MPlaces = append(pn.Net.FinalMarking.MPlaces, *mp)
}

func (pn *PNML) GenerateInvariant() string {
	ret := "!("
	first := true
	for _, mp := range pn.Net.FinalMarking.MPlaces {
		if mp.TokenCount != "0" {
			if !first {
				ret += " && "
				first = false
			} else {
				first = false
			}
			ret += fmt.Sprintf("%s==%s", mp.ID, mp.TokenCount)
		}
	}
	ret += ")"
	return ret
}
