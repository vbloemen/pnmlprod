package main

import (
	"fmt"
	"strconv"
)

type MarkingGraph struct {
	Markings []MGMarking
	Edges    []MGEdge
}

type MGPlace struct {
	ID string
}

type MGEdge struct {
	Source int
	Target int
	ID     string
	Label  string
	Type   string
}

var MGMarkingIDCount = 0

type MGMarking struct {
	ID     int
	Places []MGPlace
	info   string
}

func (pn *PNML) CreateMarkingGraph() MarkingGraph {
	mg := MarkingGraph{}
	// initial marking
	InitMarking := MGMarking{info: "init", ID: MGMarkingIDCount}
	MGMarkingIDCount += 1
	for _, place := range pn.Net.Page.Places {
		count, _ := strconv.Atoi(place.InitialMarking)
		for ; count > 0; count -= 1 {
			InitMarking.Add(place)
		}
	}
	mg.Markings = append(mg.Markings, InitMarking)

	V := []MGMarking{InitMarking}
	Q := []MGMarking{InitMarking}
	for len(Q) > 0 {
		M := Q[0] // current marking
		Q = Q[1:]
		for _, trans := range pn.Net.Page.Transitions {
			if pn.CanFire(trans, M) {
				newM := pn.Fire(trans, M)
				// check if marking is already visited
				found := false
				targetID := newM.ID
				for _, vm := range V {
					if markingEquals(vm, newM) {
						found = true
						targetID = vm.ID
					}
				}
				if !found {
					// add it to the lists
					Q = append(Q, newM)
					V = append(V, newM)
					mg.Markings = append(mg.Markings, newM)
				}
				mg.Edges = append(mg.Edges,
					MGEdge{
						ID:     trans.ID,
						Label:  trans.OrigName,
						Type:   trans.Type,
						Source: M.ID,
						Target: targetID})
			}
		}
	}

	return mg
}

func markingEquals(a, b MGMarking) bool {
	if len(a.Places) != len(b.Places) {
		return false
	}
	for _, place := range a.Places {
		// check how many are the same
		countA := 0
		countB := 0
		for _, aplace := range a.Places {
			if aplace.ID == place.ID {
				countA += 1
			}
		}
		for _, bplace := range b.Places {
			if bplace.ID == place.ID {
				countB += 1
			}
		}
		if countA != countB {
			return false
		}
	}
	return true
}

func (pn *PNML) CanFire(trans Transition, m MGMarking) bool {
	// check the in-arcs
	for _, arc := range pn.Net.Page.Arcs {
		if arc.Target == trans.ID {
			found := false
			for _, place := range m.Places {
				if arc.Source == place.ID {
					found = true
				}
			}
			if !found {
				return false
			}
		}
	}
	return true
}

func (pn *PNML) Fire(trans Transition, m MGMarking) MGMarking {
	newM := MGMarking{ID: MGMarkingIDCount}
	MGMarkingIDCount += 1
	for _, place := range m.Places {
		newM.Places = append(newM.Places, MGPlace{ID: place.ID})
	}

	// remove places
	for _, arc := range pn.Net.Page.Arcs {
		if arc.Target == trans.ID {
			found := -1
			for i, place := range newM.Places {
				if arc.Source == place.ID {
					found = i
				}
			}
			if found != -1 {
				newM.Places = append(newM.Places[:found], newM.Places[found+1:]...)
			}
		}
	}
	// add places
	for _, arc := range pn.Net.Page.Arcs {
		if arc.Source == trans.ID {
			newM.Places = append(newM.Places, MGPlace{ID: arc.Target})
		}
	}
	return newM
}

func (m *MGMarking) Add(place Place) {
	m.Places = append(m.Places, MGPlace{ID: place.ID})
}

func (m *MGMarking) Print() string {
	ret := ""
	for _, place := range m.Places {
		ret += place.ID
	}
	return ret
}

func (edge *MGEdge) dotColor() string {
	// search for the transition
	switch edge.Type {
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
	return "black"
}
func (mg *MarkingGraph) PrintDOT(filename string) {
	ret := "digraph g {\n"
	ret += "  rankdir=\"LR\";\n" // horizontal layout
	// NB: add initial/final marking info?
	for _, marking := range mg.Markings {
		ret += fmt.Sprintf("  m%d [label=\"%s\",shape=box, "+
			"style=\"filled,solid,rounded\", fillcolor=\"slategray1\", "+
			"fontname=\"Courier-Bold\"];\n",
			marking.ID, " ")
	}
	for _, edge := range mg.Edges {
		ret += fmt.Sprintf("  m%d -> m%d [label=\"%s\", penwidth=2, color=\"%s\""+
			", fontcolor=\"%s\"];\n",
			edge.Source, edge.Target, edge.Label, edge.dotColor(), edge.dotColor())
	}

	ret += "}\n"
	WriteFile(filename, ret)
}
