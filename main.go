package main

import (
	"fmt"
	"log"
	"os"
	"strings"
)

var check *CheckBreak

func main() {
	args := os.Args[1:]
	if 3 > len(args) {
		log.Fatal("Not enough arguments")
	}
	workingPath := args[0]
	startPoint := args[1]
	endPoint := args[2]
	cb, errInit := check.init(workingPath, startPoint, endPoint)
	if errInit != nil {
		log.Fatal("Init failed : ", errInit)
	}
	displayTitle(cb)
	report, errReport := cb.report()
	if errReport != nil {
		log.Fatal("Error during report construction : ", errReport)
	}
	displayBreaks(report)
	displayIgnored(report)
}

func displayTitle(cb *CheckBreak) {
	title := "\nCheck-break report : " + cb.workingPath + " " + cb.startPoint + " " + cb.endPoint
	fmt.Println(strings.Repeat("#", len(title)), title)
	fmt.Println("For details, see https://github.com/Prytoegrian/check-break#what-is-a-compatibility-break-")
	fmt.Println()
}

func displayBreaks(report *BreakReport) {
	if 0 == len(report.supported) {
		fmt.Println("> No compatibility break")
	} else {
		potential := "> Potentials compatibility breaks"
		fmt.Println(potential, "\n", strings.Repeat("-", len(potential)))

		for _, fileReport := range report.supported {
			fmt.Println(fileReport.report())
		}
	}
}

func displayIgnored(report *BreakReport) {
	if 0 != len(report.ignored) {
		ignored := "\n> Unsupported files"
		fmt.Println(ignored, "\n", strings.Repeat("-", len(ignored)))
		for _, fileIgnored := range report.ignored {
			fmt.Println(fileIgnored.report())
		}
	}
}
