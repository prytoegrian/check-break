package main

import (
	"fmt"
	"os"
	"strings"
)

var check *checkBreak

func main() {
	args := os.Args[1:]
	if 2 > len(args) {
		fmt.Println("Not enough arguments")
		os.Exit(1)
		return
	}
	path := args[0]
	baseBranch := args[1]
	workingBranch := args[2]
	cb := check.init(path, baseBranch, workingBranch)
	breakReport, errReport := cb.report()
	if nil != errReport {
		fmt.Println(errReport)
		os.Exit(1)
	}
	title := "\nCheck-break report : " + cb.path + " " + cb.delta
	fmt.Println(strings.Repeat("#", len(title)), strings.ToTitle(title))
	fmt.Println()
	if 0 == len(breakReport.files) {
		fmt.Println("> Aucun cassage de compatibilité")
		return
	}
	potential := "> Cassages de compatibilité potentiels"
	fmt.Println(potential, "\n", strings.Repeat("-", len(potential)))

	for _, fileReport := range breakReport.files {
		fileReport.report()
	}
}
