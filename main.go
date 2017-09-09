package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/prytoegrian/check-break/check"
)

func main() {
	args := os.Args[1:]
	if 3 > len(args) {
		log.Fatal("Not enough arguments")
	}
	workingPath := args[0]
	startPoint := args[1]
	endPoint := args[2]
	cb := check.New()
	cb, errInit := cb.Init(workingPath, startPoint, endPoint)
	if errInit != nil {
		log.Fatal("Init failed : ", errInit)
	}
	displayTitle(cb)
	report, errReport := cb.Report()
	if errReport != nil {
		log.Fatal("Error during report construction : ", errReport)
	}
	displayBreaks(report)
	displayIgnored(report)
}

func displayTitle(cb *check.CheckBreak) {
	title := "\nCheck-break report : " + cb.WorkingPath + " " + cb.StartPoint + " " + cb.EndPoint
	fmt.Println(strings.Repeat("#", len(title)), title)
	fmt.Println("For details, see https://github.com/Prytoegrian/check-break#what-is-a-compatibility-break-")
	fmt.Println()
}

func displayBreaks(report *check.BreakReport) {
	if 0 == len(report.Supported) {
		fmt.Println("> No compatibility break")
	} else {
		potential := "> Potentials compatibility breaks"
		fmt.Println(potential, "\n", strings.Repeat("-", len(potential)))

		for _, fileReport := range report.Supported {
			fmt.Println(fileReport.Report())
		}
	}
}

func displayIgnored(report *check.BreakReport) {
	if 0 != len(report.Ignored) {
		ignored := "\n> Unsupported files"
		fmt.Println(ignored, "\n", strings.Repeat("-", len(ignored)))
		for _, fileIgnored := range report.Ignored {
			fmt.Println(fileIgnored.Report())
		}
	}
}
