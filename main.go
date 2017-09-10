package main

import (
	"fmt"
	"log"
	"os"

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
	b, errInit := check.Init(workingPath, startPoint, endPoint)
	if errInit != nil {
		log.Fatal("Init failed : ", errInit)
	}
	displayTitle(b)
	report, errReport := b.Report()
	if errReport != nil {
		log.Fatal("Error during report construction : ", errReport)
	}
	displayBreaks(report)
	displayIgnored(report)
}

func displayTitle(b *check.Break) {
	fmt.Println("Check-break report")
	fmt.Println("(For details, please consult https://github.com/Prytoegrian/check-break#what-is-a-compatibility-break-)")
	fmt.Println()
}

func displayBreaks(report *check.BreakReport) {
	if 0 == len(report.Supported) {
		fmt.Println("> No compatibility break")
	} else {
		fmt.Println("> Potentials compatibility breaks")
		for _, fileReport := range report.Supported {
			fmt.Println(fileReport.Report())
		}
	}
}

func displayIgnored(report *check.BreakReport) {
	if 0 != len(report.Ignored) {
		fmt.Println("\n> Unsupported files :")
		for _, fileIgnored := range report.Ignored {
			fmt.Println(fileIgnored.Report())
		}
	}
}
