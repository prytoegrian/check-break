package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/prytoegrian/check-break/check"
)

func main() {
	path := flag.String("p", "", "Path to analyse (optional)")
	startingPoint := flag.String("s", "", "Git starting point")
	endingPoint := flag.String("e", "", "Git ending point")
	configFilename := flag.String("c", "cb-config.json", "Config filename, relative to analysed path (optional)")
	flag.Parse()
	if *startingPoint == "" {
		log.Fatalln("Starting point is missing, use -h for details")
	}
	if *endingPoint == "" {
		log.Fatalln("Ending point is missing, use -h for details")
	}
	b, errInit := check.Init(workingPath(*path), *startingPoint, *endingPoint, *configFilename)
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
	displayExclusions(report)
}

func workingPath(userPath string) string {
	if userPath != "" {
		return userPath
	}
	path, err := os.Getwd()
	if err != nil {
		log.Fatalln(err, path)
	}
	return strings.TrimSpace(path)
}

func displayTitle(b *check.Break) {
	fmt.Println("(For details, please consult https://github.com/Prytoegrian/check-break#what-is-a-compatibility-break-)")
	if !b.HasConfiguration() {
		fmt.Println("No config file found, checking without one")
	} else {
		fmt.Println("Using a config file")
	}
	fmt.Println()
}

func displayBreaks(report *check.BreakReport) {
	if 0 == len(report.Supported) {
		fmt.Println("> No compatibility break")
		fmt.Println()
	} else {
		fmt.Println("> Potentials compatibility breaks")
		for _, fileReport := range report.Supported {
			fmt.Println(fileReport.Report())
		}
	}
}

func displayIgnored(report *check.BreakReport) {
	if 0 != len(report.Ignored) {
		fmt.Println("> Unsupported files :")
		for _, fileIgnored := range report.Ignored {
			fmt.Println(fileIgnored.Report())
		}
	}
}

func displayExclusions(r *check.BreakReport) {
	if 0 != len(r.Exclusions) {
		fmt.Println("\n> Excluded paths :")
		for _, e := range r.Exclusions {
			fmt.Println(">>", e)
		}
	}
}
