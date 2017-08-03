package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ferhatelmas/levenshtein"
)

const (
	// MaximumSimilarity represents how close two strings has to be to be considered a change on the same line
	MaximumSimilarity = 20
)

var check *checkBreak

func main() {
	//args := os.Args[1:]
	// check error args

	cb, errInit := check.init("/home/rlecault/Gesica", "topic/optimListFf...topic/optimListFf4Real")
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
	fmt.Println("/!\\ Cassage de compatibilité :")
	for _, fileReport := range breakReport.files {
		fmt.Println("> ", fileReport.filename)
		//fmt.Println(">> Suppression :")
		for _, method := range fileReport.methods {
			fmt.Println(method)
		}
		fmt.Println("")
	}
	return

	//for _, fileLine := range files {

	//breaks := getBreaks(fileLine, diff)
	//report := FileReport{
	//	filename: file.name,
	//	methods:  getMethods(file),
	//
	//reports = append(reports, report)

	//methods := file.methods(file.diff)

	/*if 0 < len(diffDeleted) {
		fmt.Println("\n> Cassage de compatibilité de", filename, ":")
		if 0 == len(diffAdded) {
			fmt.Println("|-> Suppression : ")
			for _, suppression := range diffDeleted {
				fmt.Println("|--> ", strings.TrimSpace(suppression[2:]))
			}
		} else {
			fmt.Println("|-> Mutation : ")
		}
	}*/
	//}
	distance := levenshtein.Dist("bb", "acaib")
	fmt.Println(distance)
}

/*
func buildReport() BreakReport {
	files := filesReport()

	return BreakReport{
		file:    file,
		methods: getMethods(file),
	}
}
*/

func (f *File) filterDiff() {
	pattern := getPattern(f.typeFile)
	deletions := f.diff.deletions
	for index, deletion := range deletions {
		if pattern.MatchString(deletion) {
			deletions = append(deletions[:index], deletions[index+1])
		}
	}
	addings := f.diff.addings
	for index, adding := range addings {
		if pattern.MatchString(adding) {
			addings = append(addings[:index], addings[index+1])
		}
	}
}

/*
func getBreaks(pathFile string, diff string) *Method {
	extension := path.Ext(pathFile)
	pattern := getPattern(extension)

	var results []string
	if nil == pattern {
		return nil
	}
	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if pattern.MatchString(line) {
			results = append(results, line)

		}
	}
	return Method{}

}
*/
func getPattern(extension string) *regexp.Regexp {
	var pattern *regexp.Regexp
	switch extension {
	case ".go":
		pattern, _ = regexp.Compile(`func [A-Za-z]*\(`)
	case ".php":
		pattern, _ = regexp.Compile(`(\s)+public( static)? function [_A-Za-z]`)
	}

	return pattern
}

func getMethods(file File) []Method {
	var methods []Method

	return methods
}
