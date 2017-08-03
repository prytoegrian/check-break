package main

import (
	"os"
	"path"
	"strings"

	"github.com/tbruyelle/qexec"
)

type checkBreak struct {
	path   string
	commit string
}

// BreakReport is a report to display
type BreakReport struct {
	files []FileReport
}

// FileReport is a filename and a slice of all potentials compatibility break
type FileReport struct {
	methods  []Method
	filename string
}

// Method is a potential break on a public method
type Method struct {
	before   string
	after    string
	distance int
}

func (*checkBreak) init(path string, commit string) (*checkBreak, error) {
	//var err error
	os.Chdir(path)

	return &checkBreak{
		path:   path,
		commit: commit,
	}, nil

	// not a git repo
	// not…
}

func (cb *checkBreak) report() (*BreakReport, error) {
	gitFiles, err := qexec.Run("git", "diff", "--name-status", cb.commit) // branche de base
	if nil != err {
		return nil, err
	}

	files := filesToReport(gitFiles, *cb)

	filesReports := make([]FileReport, 0)
	for _, file := range files {
		fileReport := FileReport{
			filename: file.name,
		}
		filesReports = append(filesReports, fileReport)
	}

	return &BreakReport{
		files: filesReports,
	}, nil
}

// File is a file representation
type File struct {
	name     string
	status   string
	diff     Diff
	typeFile string
}

// Diff represents the diff of a file, segregated with deletion and adding
type Diff struct {
	deletions []string
	addings   []string
}

func filesToReport(filenamesDiff string, cb checkBreak) []File {
	files := make([]File, 0)

	for _, fileLine := range strings.Split(strings.TrimSpace(filenamesDiff), "\n") {
		if !isUsefulFile(fileLine) {
			continue
		}

		file := &File{}
		file.setStatus(fileLine)
		file.setName(fileLine)
		file.setType(fileLine)
		file.setDiff(cb)

		if isTypeDefined(*file) {
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
	f.typeFile = path.Ext(fileLine)
}

func (f *File) setDiff(cb checkBreak) error {
	diff, err := qexec.Run("git", "diff", cb.commit, f.name) // branche de base
	if nil != err {
		return err
	}

	var diffDeleted []string
	var diffAdded []string
	for _, line := range strings.Split(diff, "\n") {
		if strings.HasPrefix(line, "- ") {
			diffDeleted = append(diffDeleted, line)
		} else if strings.HasPrefix(line, "+ ") {
			diffAdded = append(diffAdded, line)
		}
	}

	f.diff = Diff{
		deletions: diffDeleted,
		addings:   diffAdded,
	}
	return nil
}

func isUsefulFile(fileLine string) bool {
	return !strings.HasPrefix(fileLine, "A")
}

func isTypeDefined(file File) bool {
	types := []string{".go", ".php"}

	for _, typeFile := range types {
		if file.typeFile == typeFile {
			return true
		}
	}
	return false
}

// construire une structure pour la récupération des BC en fonction des langages
