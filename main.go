package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/prytoegrian/check-break/check"
)

func main() {
	path := flag.String("p", "", "Path to analyse")
	startingPoint := flag.String("s", "", "Git starting point")
	endingPoint := flag.String("e", "", "Git ending point")
	configPath := flag.String("c", "", "Config file path (optional)")
	flag.Parse()
	if *path == "" {
		log.Fatalln("Path is missing")
	}
	if *startingPoint == "" {
		log.Fatalln("Starting point is missing")
	}
	if *endingPoint == "" {
		log.Fatalln("Ending point is missing")
	}
	b, errInit := check.Init(*path, *startingPoint, *endingPoint, *configPath)
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
