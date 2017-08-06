package main

import (
	"fmt"
	"os"
	"strings"
)

var check *CheckBreak

func main() {
	args := os.Args[1:]
	if 3 > len(args) {
		fmt.Println("Not enough arguments")
		os.Exit(1)
		return
	}
	path := args[0]
	startPoint := args[1]
	endPoint := args[2]
	cb, errInit := check.init(path, startPoint, endPoint)
	if nil != errInit {
		fmt.Println("Initialisation impossible :", errInit)
		os.Exit(1)
	}
	report, errReport := cb.report()
	if nil != errReport {
		fmt.Println("Erreur durant la construction du rapport :", errReport)
		os.Exit(1)
	}
	title := "\nCheck-break report : " + cb.path + " " + cb.startPoint + " " + cb.endPoint
	fmt.Println(strings.Repeat("#", len(title)), strings.ToTitle(title))
	fmt.Println()
	if 0 == len(report.supported) {
		fmt.Println("> Aucun cassage de compatibilité")
		return
	}
	potential := "> Cassages de compatibilité potentiels"
	fmt.Println(potential, "\n", strings.Repeat("-", len(potential)))

	for _, fileReport := range report.supported {
		fmt.Println(fileReport.report())
	}

	if 0 != len(report.ignored) {
		ignored := "> Fichiers ignorés"
		fmt.Println(ignored, "\n", strings.Repeat("-", len(ignored)))
		for _, fileIgnored := range report.ignored {
			fmt.Println(fileIgnored.report())
		}
	}
}
