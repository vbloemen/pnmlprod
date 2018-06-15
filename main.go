/* Parse a PNML model and log trace, then produce PNML of synchronous product
 * while annotating special actions (tau,sync,log,model)
 */
package main

import (
	"fmt"
	"os"
)

func showHelp() {
	// TODO: provide output dir?
	fmt.Println("USAGE:")
	fmt.Printf("    %v  -p  MODEL.pnml  LOGFILE.{csv,xes}  OUTPUTDIR\n",
		os.Args[0])
	fmt.Printf("\n")
	fmt.Printf("        %s\n", "Constructs a synchronous product"+
		" for each log trace in LOGFILE.xes, to be\n        used in"+
		"pnml2lts-sym for computing an alignment trace.\n        "+
		"Two files are created per log trace x: 'syncmodel-x.pnml' and"+
		" 'invariant-x.txt'")
	fmt.Printf("\n")
	fmt.Printf("    %v  -a  SYNCMODEL.pnml  TRACE.txt\n", os.Args[0])
	fmt.Printf("\n")
	fmt.Printf("        %s\n", "Constructs an alignment from the synchronous"+
		" product and the log trace.\n        "+
		"The resulting alignment is printed on the standard output")
	fmt.Printf("\n")
	fmt.Printf("    %v  -c  MODEL.pnml\n", os.Args[0])
	fmt.Printf("\n")
	fmt.Printf("        %s\n", "Returns the size of the Petri net model; the"+
		" number of places, transitions and arcs")
	//		"\n        A DOT file 'SYNCMODEL.dot' is also constructed")
	os.Exit(0)
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func WriteFile(filename, contents string) {
	file, err := os.Create(filename)
	CheckError(err)
	defer file.Close()
	_, err = file.WriteString(contents)
	CheckError(err)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: insufficient arguments")
		showHelp()
	}
	if os.Args[1] != "-a" && os.Args[1] != "-p" && os.Args[1] != "-c" {
		fmt.Println("Error: unknown option: '" + os.Args[1] + "'")
		showHelp()
	} else if os.Args[1] == "-p" {
		if len(os.Args) != 5 {
			fmt.Println("Error: insufficient arguments")
			showHelp()
		}
		CreatePNMLProduct(os.Args[2], os.Args[3], os.Args[4])
	} else if os.Args[1] == "-a" {
		if len(os.Args) != 4 {
			fmt.Println("Error: insufficient arguments")
			showHelp()
		}
		TraceToAlign(os.Args[2], os.Args[3])
	} else if os.Args[1] == "-c" {
		if len(os.Args) != 3 {
			fmt.Println("Error: insufficient arguments")
			showHelp()
		}
		CheckModel(os.Args[2])
	}
}
