package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type TracePart struct {
	MoveType    string
	PlaceIDs    []string
	PlaceTokens []int
}

const (
	SKIP string = "»"
)

var (
	Trmap map[string]Transition
	Trace []TracePart
)

// - "SYNC-[in,in,in]-[out,out,out]" -> transition
// NB: we might return a list of transitions, but a single one is enough
func (pn *PNML) CreateTransitionMap() {
	Trmap = make(map[string]Transition)
	for _, trans := range pn.Net.Page.Transitions {
		var in, out []string
		// find the corresponding arcs
		for _, arc := range pn.Net.Page.Arcs {
			if arc.Target == trans.ID {
				in = append(in, arc.Source)
			}
			if arc.Source == trans.ID {
				out = append(out, arc.Target)
			}
		}
		sort.Strings(in)
		sort.Strings(out)
		Trmap[fmt.Sprintf("%s-%s-%s", trans.Type, in, out)] = trans
	}
}

func AddAlignPair(t Transition) {
	pair := AlignPair{Log: SKIP, Trans: SKIP, TransID: ""}
	if t.Type == LOG || t.Type == SYNC {
		pair.Log = t.OrigName
		pair.TransID = t.ID
	}
	if t.Type == MODEL || t.Type == SYNC || t.Type == TAU {
		pair.Trans = t.OrigName
		pair.TransID = t.ID
	}
	Alignment.Pairs = append(Alignment.Pairs, pair)
}

func TraceToAlignOld(syncmodelfn, tracefn string) {
	// read trace
	file, err := os.Open(tracefn)
	CheckError(err)
	defer file.Close()

	// read PNML
	modelcontents := readPNML(syncmodelfn)
	var pn PNML
	xml.Unmarshal(modelcontents, &pn) // fill in PNML contents

	// initialize mapping from string to transitions (there might be some loss
	// of information, but this shouldn't be a problem)
	pn.CreateTransitionMap()

	// put the information in TraceParts, and collect these in Trace
	var currentTP TracePart = TracePart{MoveType: "INITIAL"}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "action") {
			Trace = append(Trace, currentTP)
			currentTP = TracePart{}
			for _, movetype := range MOVES {
				if strings.Contains(line, movetype) {
					currentTP.MoveType = movetype
				}
			}
		} else if strings.Contains(line, "place") {
			split := strings.Split(line, " ")
			if len(split) < 3 {
				CheckError(errors.New("Unable to parse line: '" + line + "'"))
			}
			placeID := strings.Split(split[len(split)-3], ":")[0]
			tokencount, err := strconv.Atoi(split[len(split)-1])
			CheckError(err)
			currentTP.PlaceIDs = append(currentTP.PlaceIDs, placeID)
			currentTP.PlaceTokens = append(currentTP.PlaceTokens, tokencount)
		}
	}
	Trace = append(Trace, currentTP)

	// Form an alignment from the Trace object
	var marking = make(map[string]int)
	for _, tp := range Trace {
		if tp.MoveType == "INITIAL" {
			for i, placeID := range tp.PlaceIDs {
				marking[placeID] = tp.PlaceTokens[i]
			}
			continue
		}
		// calculate difference and update the markings
		var inP, outP []string
		for i, placeID := range tp.PlaceIDs {
			tokendiff := tp.PlaceTokens[i] - marking[placeID]
			if tokendiff < 0 {
				for i := tokendiff; i < 0; i++ {
					inP = append(inP, placeID)
				}
			} else if tokendiff > 0 {
				for i := 0; i < tokendiff; i++ {
					outP = append(outP, placeID)
				}
			}
			marking[placeID] = tp.PlaceTokens[i]
		}

		sort.Strings(inP)
		sort.Strings(outP)
		trans := Trmap[fmt.Sprintf("%s-%s-%s", tp.MoveType, inP, outP)]
		fmt.Printf("%s in: %s out: %s, %v\n\n", tp.PlaceIDs, inP, outP, trans)
		if trans.Type == "" {
			CheckError(errors.New(fmt.Sprintf("Unable to form transition from"+
				" marking difference: %s-%s-%s. Are you sure the input files"+
				" are up to date?", tp.MoveType, inP, outP)))
		}
		AddAlignPair(trans)
		// change type in actual transition
		for ti, tr := range pn.Net.Page.Transitions {
			if tr.ID == trans.ID {
				pn.Net.Page.Transitions[ti].Selected = "true"
			}
		}
	}

	fmt.Println(Alignment.toString())
	//pn.PrintDOT(syncmodelfn[0:len(syncmodelfn)-5] + ".dot")
}

// new trace to align where entire markings are compared, to avoid problems
// like (place) <--> [trans]

type Trans struct {
	T    Transition
	Type string
	In   []int
	Out  []int
}

type TransArr struct {
	Trans []Trans
}

func (pn *PNML) MakeTransArr(placeMap map[string]int) TransArr {
	ret := TransArr{}
	ret.Trans = make([]Trans, len(pn.Net.Page.Transitions))

	for ti, trans := range pn.Net.Page.Transitions {
		T := Trans{}
		T.Type = trans.Type
		T.T = trans
		var in, out []string // store the places
		// find the corresponding arcs
		for _, arc := range pn.Net.Page.Arcs {
			if arc.Target == trans.ID {
				in = append(in, arc.Source)
			}
			if arc.Source == trans.ID {
				out = append(out, arc.Target)
			}
		}
		sort.Strings(in)
		sort.Strings(out)
		T.In = make([]int, len(in))
		T.Out = make([]int, len(out))
		for i, place := range in {
			T.In[i] = placeMap[place]
		}
		for i, place := range out {
			T.Out[i] = placeMap[place]
		}
		ret.Trans[ti] = T
	}

	return ret
}

func TraceToAlign(syncmodelfn, tracefn string) {
	// read trace
	file, err := os.Open(tracefn)
	CheckError(err)
	defer file.Close()

	// read PNML
	modelcontents := readPNML(syncmodelfn)
	var pn PNML
	xml.Unmarshal(modelcontents, &pn) // fill in PNML contents

	// put the information in TraceParts, and collect these in Trace
	var currentTP TracePart = TracePart{MoveType: "INITIAL"}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "action") {
			Trace = append(Trace, currentTP)
			currentTP = TracePart{}
			for _, movetype := range MOVES {
				if strings.Contains(line, movetype) {
					currentTP.MoveType = movetype
				}
			}
		} else if strings.Contains(line, "place") {
			split := strings.Split(line, " ")
			if len(split) < 3 {
				CheckError(errors.New("Unable to parse line: '" + line + "'"))
			}
			placeID := strings.Split(split[len(split)-3], ":")[0]
			tokencount, err := strconv.Atoi(split[len(split)-1])
			CheckError(err)
			currentTP.PlaceIDs = append(currentTP.PlaceIDs, placeID)
			currentTP.PlaceTokens = append(currentTP.PlaceTokens, tokencount)
		}
	}
	Trace = append(Trace, currentTP)

	// create mapping for placeArr
	placeMap := make(map[string]int)
	for i, p := range Trace[0].PlaceIDs {
		placeMap[p] = i
	}

	// create TransArr
	Tarr := pn.MakeTransArr(placeMap)

	// set initial marking
	L := len(Trace[0].PlaceTokens)
	currentMarking := make([]int, L)
	newMarking := make([]int, L)
	tmpMarking := make([]int, L)
	for i, _ := range Trace[0].PlaceTokens {
		currentMarking[i] = Trace[0].PlaceTokens[i]
		newMarking[i] = Trace[0].PlaceTokens[i]
	}

	// search for matching transition
	for _, tp := range Trace {
		if tp.MoveType == "INITIAL" {
			continue
		}
		// set new marking
		for i, place := range tp.PlaceIDs {
			newMarking[placeMap[place]] = tp.PlaceTokens[i]
		}

		// search for transitions that can be fired on the current marking
		foundTrans := false
		for _, trans := range Tarr.Trans {
			if trans.Type != tp.MoveType {
				continue
			}
			// check in transitions
			canfire := true
			for _, in := range trans.In {
				if currentMarking[in] <= 0 {
					canfire = false
					break
				}
			}
			if !canfire {
				continue
			}

			// check if the out transitions
			//fmt.Println(trans)
			for i, n := range currentMarking {
				tmpMarking[i] = n
			}
			for _, in := range trans.In {
				tmpMarking[in] -= 1
				if tmpMarking[in] < 0 { // in case multiple tokens are subtracted
					canfire = false
					break
				}
			}
			if !canfire {
				continue
			}
			// add outgoing tokens
			for _, out := range trans.Out {
				tmpMarking[out] += 1
			}
			// compare markings
			for i, n := range newMarking {
				if tmpMarking[i] != n {
					canfire = false
					break
				}
			}
			if canfire { // correct transition chosen!
				AddAlignPair(trans.T)
				foundTrans = true
				break
			}
		}
		if !foundTrans {
			CheckError(errors.New(
				fmt.Sprintf("Could not find fitting transition: %v", tp)))
		}

		for i, n := range newMarking {
			currentMarking[i] = n
		}
	}

	fmt.Println(Alignment.toString())
}
