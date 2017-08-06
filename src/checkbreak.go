package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/ferhatelmas/levenshtein"
	"github.com/tbruyelle/git"
	"github.com/tbruyelle/qexec"
)

type CheckBreak struct {
	path       string
	startPoint string
	endPoint   string
}

func (cb *CheckBreak) init(path string, startPoint string, endPoint string) (*CheckBreak, error) {
	if errPath := os.Chdir(path); nil != errPath {
		return nil, errors.New("Path inexistant")
	}

	if exists, _ := git.RefExists(startPoint); !exists {
		return nil, errors.New("L'objet « " + startPoint + " » n'existe pas")
	}

	if exists, _ := git.RefExists(endPoint); !exists {
		return nil, errors.New("L'objet « " + endPoint + " » n'existe pas")
	}

	return &CheckBreak{
		path:       path,
		startPoint: startPoint,
		endPoint:   endPoint,
	}, nil
}

// BreakReport is a report to display
type BreakReport struct {
	supported []FileReport
	ignored   []File
}

// report displays a BreakReport
func (cb *CheckBreak) report() (*BreakReport, error) {
	gitFiles, err := qexec.Run("git", "diff", "--name-status", cb.startPoint+"..."+cb.endPoint)
	if nil != err {
		return nil, err
	}

	supported, ignored := files(gitFiles, *cb)

	filesReports := make([]FileReport, 0)
	for _, file := range supported {
		methods, _ := file.breaks()

		if 0 != len(*methods) {
			fileReport := FileReport{
				filename: file.name,
				methods:  *methods,
			}
			filesReports = append(filesReports, fileReport)
		}
	}

	return &BreakReport{
		supported: filesReports,
		ignored:   ignored,
	}, nil
}

// FileReport is a pool of potentials compatibility breaks
type FileReport struct {
	methods  []Method
	filename string
}

// report displays a FileReport and its potentials compatibility breaks
func (fr *FileReport) report() string {
	report := ">> " + fr.filename
	for _, method := range fr.methods {
		report += "\n"
		beforeFormatted := color.RedString("-" + method.before)
		afterFormatted := color.GreenString("+" + method.after)
		if "" == method.after {
			report += beforeFormatted
		} else {
			report += beforeFormatted + " -> " + afterFormatted
		}
	}

	return report + "\n"
}

// File is a file representation
type File struct {
	name     string
	status   string
	diff     Diff
	typeFile string
}

// Method is a potential break on a public method
type Method struct {
	before       string
	after        string
	commonFactor string
	distance     int
}

func (f *File) report() string {
	return fmt.Sprint(">> ", color.CyanString(f.name))
}

// breaks returns all potentials CB on a file
func (f *File) breaks() (*[]Method, error) {
	pattern, err := f.breakPattern()
	if nil != err {
		return nil, err
	}

	var methods []Method
	var moveOnly bool
	for _, deleted := range f.diff.deletions {
		var closestAdding string
		moveOnly = false
		commonFactor := pattern.FindStringSubmatch(deleted)[0]
		for _, added := range f.diff.addings {
			if strings.HasPrefix(added, commonFactor) {
				// It's only a move
				if len(deleted) == len(added) {
					moveOnly = true
					break
				} else {
					closestAdding = added
				}
			}
		}
		if !moveOnly {
			method := Method{
				before:       deleted,
				after:        closestAdding,
				commonFactor: commonFactor,
				distance:     levenshtein.Dist(deleted, closestAdding),
			}
			methods = append(methods, method)
		}
	}

	return &methods, nil
}

// files initializes files struct
func files(filenamesDiff string, cb CheckBreak) ([]File, []File) {
	supported := make([]File, 0)
	ignored := make([]File, 0)

	for _, fileLine := range strings.Split(strings.TrimSpace(filenamesDiff), "\n") {
		file := File{}
		status, name, filetype := extractDataFile(fileLine)
		file.name = name
		file.status = status
		file.typeFile = filetype
		diff, err := file.getDiff(cb.startPoint, cb.endPoint)
		if nil == err {
			file.diff = *diff
		}

		if file.canHaveBreak() {
			if file.isTypeSupported() {
				supported = append(supported, file)
			} else {
				ignored = append(ignored, file)
			}
		}
	}

	return supported, ignored
}

// extractDataFile gives file's name, status and type
func extractDataFile(fileLine string) (string, string, string) {
	fields := strings.Fields(fileLine)
	status := fields[0]
	name := fields[1]
	if strings.HasPrefix(status, "R") {
		status = "D"
	}

	return status, name, typefile(name)
}

// typeFile return a file's extension
func typefile(filepath string) string {
	var typeFile string
	filename := path.Base(filepath)
	if strings.Contains(filename, ".") && !strings.HasPrefix(filename, ".") {
		typeFile = strings.TrimSpace(path.Ext(filename)[1:])
	}

	return typeFile
}

func (f *File) canHaveBreak() bool {
	return "A" != f.status
}

func (f *File) isDeleted() bool {
	return "D" == f.status
}

// Diff represents the diff of a file, segregated with deletion and adding
type Diff struct {
	deletions []string
	addings   []string
}

func (f *File) getDiff(startObject string, endObject string) (*Diff, error) {
	if f.isDeleted() {
		return f.getDiffDeleted(startObject)
	}
	diff, err := qexec.Run("git", "diff", "-U0", startObject+"..."+endObject, f.name)
	if nil != err {
		return nil, err
	}

	pattern, errPattern := f.breakPattern()
	if nil != errPattern {
		return nil, errPattern
	}

	var diffDeleted []string
	var diffAdded []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "-") {
			diffDeleted = append(diffDeleted, strings.TrimSpace(line[1:]))
		} else if strings.HasPrefix(line, "+") {
			diffAdded = append(diffAdded, strings.TrimSpace(line[1:]))
		}
	}

	return &Diff{
		deletions: filteredByPattern(pattern, diffDeleted),
		addings:   filteredByPattern(pattern, diffAdded),
	}, nil
}

func (f *File) getDiffDeleted(startObject string) (*Diff, error) {
	fileDeleted, err := qexec.Run("git", "show", startObject+":"+f.name)

	if nil != err {
		return nil, err
	}
	pattern, errPattern := f.breakPattern()
	if nil != errPattern {
		return nil, errPattern
	}
	var diffDeleted []string
	for _, line := range strings.Split(fileDeleted, "\n") {
		diffDeleted = append(diffDeleted, strings.TrimSpace(line))
	}

	return &Diff{
		deletions: filteredByPattern(pattern, diffDeleted),
	}, nil
}

// filteredByPattern keeps only data lines that match a pattern
func filteredByPattern(r *regexp.Regexp, data []string) []string {
	filtered := make([]string, 0)
	for _, element := range data {
		if r.MatchString(element) {
			filtered = append(filtered, element)
		}
	}

	return filtered
}

func (f *File) isTypeSupported() bool {
	_, err := f.breakPattern()
	if nil != err {
		return false
	}
	return true
}

// breakPattern returns the regex of a potential compatibility break associated
// with type of the file
func (f *File) breakPattern() (*regexp.Regexp, error) {
	var pattern *regexp.Regexp
	switch f.typeFile {
	case "go":
		pattern = regexp.MustCompile(`^(\s)*func( \(.+)\)? [A-Z]{1}[A-Za-z]*\(`)
	case "php":
		pattern = regexp.MustCompile(`^(\s)*public( static)? function [_A-Za-z]+\(`)
	case "java":
		pattern = regexp.MustCompile(`^(\s)*public( static)?( .+)? [A-Za-z]+\(`)
	}

	if nil == pattern {
		return pattern, errors.New("Langage inconnu")
	}

	return pattern, nil
}
