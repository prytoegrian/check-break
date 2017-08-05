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
	files []FileReport
}

// report displays a BreakReport
func (cb *checkBreak) report() (*BreakReport, error) {
	gitFiles, err := qexec.Run("git", "diff", "--name-status", cb.delta)
	if nil != err {
		return nil, err
	}

	files := files(gitFiles, *cb)

	filesReports := make([]FileReport, 0)
	for _, file := range files {
		if !file.isUsefulFile() {
			continue
		}

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
		files: filesReports,
	}, nil
}

// FileReport is a pool of potentials compatibilities breaks
type FileReport struct {
	methods  []Method
	filename string
}

// report displays a FileReport and its potentials compatibility breaks
func (fr *FileReport) report() {
	fmt.Println(">>", fr.filename)
	for _, method := range fr.methods {
		beforeFormatted := color.RedString("-" + method.before)
		afterFormatted := color.GreenString("+" + method.after)
		if "" == method.after {
			fmt.Println(beforeFormatted)
		} else {
			fmt.Println(beforeFormatted, "->", afterFormatted)
		}
	}
	fmt.Println("")
}

// File is a file representation
type File struct {
	name     string
	status   string
	diff     Diff
	typeFile string
}

// isUsefulFile checks if the file will be elligible to CB
func (f *File) isUsefulFile() bool {
	return "A" != f.status
}

// Method is a potential break on a public method
type Method struct {
	before       string
	after        string
	commonFactor string
	distance     int
}

// breaks returns all potentials CB on a file
func (f *File) breaks() (*[]Method, error) {
	f.filterDiff()
	pattern, err := patternByLanguage(f.typeFile)
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

// filterDiff filters out git diff lines which aren't relevant for BC
func (f *File) filterDiff() error {
	pattern, err := patternByLanguage(f.typeFile)
	if nil != err {
		return err
	}
	f.diff = Diff{
		deletions: filteredByPattern(pattern, f.diff.deletions),
		addings:   filteredByPattern(pattern, f.diff.addings),
	}

	return nil
}

// filteredByPattern keeps only data lines which match a pattern
func filteredByPattern(r *regexp.Regexp, data []string) []string {
	filtered := make([]string, 0)
	for _, element := range data {
		if r.MatchString(element) {
			filtered = append(filtered, element)
		}
	}

	return filtered
}

// files initiliazes files struct
func files(filenamesDiff string, cb checkBreak) []File {
	files := make([]File, 0)

	for _, fileLine := range strings.Split(strings.TrimSpace(filenamesDiff), "\n") {
		file := &File{}
		file.setStatus(fileLine)
		file.setName(fileLine)
		file.setType(fileLine)
		file.setDiff(cb)

		if file.hasAutorizedType() {
			files = append(files, *file)
		}
	}

	return files
}

func (f *File) setStatus(fileLine string) {
	f.status = strings.TrimSpace(fileLine[:1])
}

func (f *File) setName(fileLine string) {
	f.name = strings.TrimSpace(fileLine[1:])
}

func (f *File) setType(fileLine string) {
	f.typeFile = strings.TrimSpace(path.Ext(fileLine)[1:])
}

// Diff represents the diff of a file, segregated with deletion and adding
type Diff struct {
	deletions []string
	addings   []string
}

func (f *File) setDiff(cb checkBreak) error {
	diff, err := qexec.Run("git", "diff", "-U0", cb.delta, f.name)
	if nil != err {
		return err
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
		deletions: diffDeleted,
		addings:   diffAdded,
	}
	return nil
}

// hasAutorizedType checks if file type is describe
func (f *File) hasAutorizedType() bool {
	_, err := patternByLanguage(f.typeFile)
	if nil != err {
		return false
	}
	return true
}

// patternByLanguage returns the regex of a potential compatibility break
func patternByLanguage(extension string) (*regexp.Regexp, error) {
	var pattern *regexp.Regexp
	switch extension {
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
