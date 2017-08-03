package main

import (
	"fmt"
	"os"
)

var check *checkBreak

func main() {
	args := os.Args[1:]
	// check error args

	cb, errInit := check.init(args[0], args[1])
	if nil != errInit {
		fmt.Println("Error :", errInit)
		os.Exit(1)
	}

	breakReport, errReport := cb.report()
	if nil != errReport {
		fmt.Println(errReport)
		os.Exit(1)
	}
	if 0 == len(breakReport.files) {
		fmt.Println("Aucun cassage de compatibilité")
		return
	}
	separator := "#####################################"
	fmt.Println(separator, "\nCassage de compatibilité potentiel")
	fmt.Println("")

	for _, fileReport := range breakReport.files {
		fileReport.report()
	}
}
