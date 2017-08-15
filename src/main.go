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
	path := args[0]
	startPoint := args[1]
	endPoint := args[2]
	cb, errInit := check.init(path, startPoint, endPoint)
	if errInit != nil {
		log.Fatal("Init impossible : ", errInit)
	}
	report, errReport := cb.report()
	if errReport != nil {
		log.Fatal("Erreur durant la construction du rapport : ", errReport)
	}
	title := "\nCheck-break report : " + cb.path + " " + cb.startPoint + " " + cb.endPoint
	fmt.Println(strings.Repeat("#", len(title)), strings.ToTitle(title))
	fmt.Println()
	if 0 == len(report.supported) {
		fmt.Println("> Aucun cassage de compatibilité")
	} else {
		potential := "> Cassages de compatibilité potentiels"
		fmt.Println(potential, "\n", strings.Repeat("-", len(potential)))

		for _, fileReport := range report.supported {
			fmt.Println(fileReport.report())
		}
	}

	if 0 != len(report.ignored) {
		ignored := "\n> Fichiers ignorés"
		fmt.Println(ignored, "\n", strings.Repeat("-", len(ignored)))
		for _, fileIgnored := range report.ignored {
			fmt.Println(fileIgnored.report())
		}
	}
}
