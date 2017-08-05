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
	"github.com/tbruyelle/qexec"
)

type checkBreak struct {
	path          string
	delta         string
	baseBranch    string
	workingBranch string
}

func (*checkBreak) init(path string, baseBranch string, workingBranch string) checkBreak {
	os.Chdir(path)

	return checkBreak{
		path:          path,
		workingBranch: workingBranch,
		baseBranch:    baseBranch,
		delta:         baseBranch + "..." + workingBranch,
	}
}

// BreakReport is a report to display
type BreakReport struct {
	supported []FileReport
	ignored   []File
}

// report displays a BreakReport
func (cb *checkBreak) report() (*BreakReport, error) {
	gitFiles, err := qexec.Run("git", "diff", "--name-status", cb.delta)
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
func files(filenamesDiff string, cb checkBreak) ([]File, []File) {
	supported := make([]File, 0)
	ignored := make([]File, 0)

	for _, fileLine := range strings.Split(strings.TrimSpace(filenamesDiff), "\n") {
		name := strings.TrimSpace(fileLine[1:])
		file := File{
			name:     name,
			status:   strings.TrimSpace(fileLine[:1]),
			typeFile: typeFile(name),
		}

		file.setFilteredDiff(cb.delta)
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

// typeFile return a file's extension
func typeFile(filepath string) string {
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

// Diff represents the diff of a file, segregated with deletion and adding
type Diff struct {
	deletions []string
	addings   []string
}

func (f *File) setFilteredDiff(deltaDiff string) error {
	diff, err := qexec.Run("git", "diff", "-U0", deltaDiff, f.name)
	if nil != err {
		return err
	}

	pattern, errPattern := f.breakPattern()
	if nil != errPattern {
		return errPattern
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
	f.diff = Diff{
		deletions: filteredByPattern(pattern, diffDeleted),
		addings:   filteredByPattern(pattern, diffAdded),
	}
	return nil
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
