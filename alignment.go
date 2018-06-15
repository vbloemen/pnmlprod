package main

import (
	"fmt"
)

var Alignment AlignmentS

type AlignmentS struct {
	Pairs []AlignPair
}

type AlignPair struct {
	Log     string
	Trans   string
	TransID string
}

func (al *AlignmentS) toString() string {
	ret := ""
	for _, pair := range al.Pairs {
		ret += pair.toString() + "\n"
	}
	return ret
}

func (ap *AlignPair) toString() string {
	return fmt.Sprintf("(%s | %s : %s)", ap.Log, ap.Trans, ap.TransID)
}
