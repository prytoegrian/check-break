package check

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/fatih/color"
	"github.com/tbruyelle/git"
	"github.com/tbruyelle/qexec"
)

// Break represents base structure required for evaluating code changes
type Break struct {
	workingPath string
	startPoint  string
	endPoint    string
	config      *config
}

// Init bootstraps Break structure
func Init(workingPath string, startPoint string, endPoint string, configPath string) (*Break, error) {
	if errPath := os.Chdir(workingPath); errPath != nil {
		return nil, fmt.Errorf("Path %s doesn't exist", workingPath)
	}

	if exists, _ := git.RefExists(startPoint); !exists {
		return nil, fmt.Errorf("The object %s doesn't exist", startPoint)
	}

	if exists, _ := git.RefExists(endPoint); !exists {
		return nil, fmt.Errorf("The object %s doesn't exist", endPoint)
	}
	config, err := loadConfiguration(configPath)
	if err != nil {
		return nil, err
	}

	return &Break{
		workingPath: workingPath,
		startPoint:  startPoint,
		endPoint:    endPoint,
		config:      config,
	}, nil
}

type config struct {
	Excluded struct {
		Path string `json:"path"`
	} `json:"excluded"`
}

func loadConfiguration(configPath string) (*config, error) {
	var conf config
	if configPath == "" {
		return &conf, nil
	}
	configFile, err := os.Open(configPath)
	defer configFile.Close()
	if err != nil {
		return nil, err
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)
	return &conf, nil
}

// BreakReport is a report to display
type BreakReport struct {
	Supported  []FileReport
	Ignored    []file
	Exclusions []string
}

// Report displays a BreakReport
func (b *Break) Report() (*BreakReport, error) {
	gitFiles, err := qexec.Run("git", "diff", "--name-status", b.startPoint+"..."+b.endPoint)
	if err != nil {
		return nil, err
	}
	f := strings.Split(strings.TrimSpace(gitFiles), "\n")
	supported, ignored := files(f, *b)
	analysables := b.filter(supported)
	ignored = b.filter(ignored)

	filesReports := make([]FileReport, 0)
	for _, file := range analysables {
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
		Supported:  filesReports,
		Ignored:    ignored,
		Exclusions: b.exclusions(),
	}, nil
}

// filter drops a file if it satisfies exclusion criteria
func (b *Break) filter(files []file) []file {
	excluded := make([]string, 0)
	if 0 == len(b.exclusions()) {
		return files
	}
	filtered := make([]file, 0)
	excluded = append(excluded, b.config.Excluded.Path)
	for _, f := range files {
		for _, e := range excluded {
			if strings.HasPrefix(f.name, e) {
				continue
			}
			filtered = append(filtered, f)
		}
	}

	return filtered
}

// exclusions is the exclusion list provided by config file
func (b *Break) exclusions() []string {
	excluded := make([]string, 0)
	if b.config.Excluded.Path != "" {
		excluded = append(excluded, b.config.Excluded.Path)
	}

	return excluded
}

// FileReport is a pool of potentials compatibility breaks
type FileReport struct {
	methods  []method
	filename string
}

// Report displays a FileReport and its potentials compatibility breaks
func (fr *FileReport) Report() string {
	report := ">> " + color.CyanString(fr.filename+" :")
	for _, method := range fr.methods {
		var change string
		report += "\n"
		beforeFormatted := color.RedString(method.before)
		afterFormatted := color.GreenString(method.after)
		if "" == method.after {
			change = beforeFormatted
		} else {
			change = beforeFormatted + " -> " + afterFormatted
		}
		report += method.explanation + " : " + change
	}

	return report + "\n"
}

// file is a file representation
type file struct {
	name     string
	status   string
	diff     diff
	typeFile string
}

// method is a potential break on a public method
type method struct {
	before       string
	after        string
	commonFactor string
	explanation  string
}

func (f *file) Report() string {
	return fmt.Sprint(">> ", color.CyanString(f.name))
}

// breaks returns all potentials CB on a file
func (f *file) breaks() (*[]method, error) {
	pattern, err := f.breakPattern()
	if err != nil {
		return nil, err
	}

	var methods []method
	var moveOnly bool
	for _, deleted := range f.diff.deletions {
		var closestAdding string
		moveOnly = false
		commonFactor := pattern.FindStringSubmatch(deleted)[0]
		for _, added := range f.diff.addings {
			if strings.HasPrefix(added, commonFactor) {
				// It's only a move
				if len(strings.Split(deleted, " ")) == len(strings.Split(added, " ")) && len(deleted) == len(added) {
					moveOnly = true
					break
				} else {
					closestAdding = added
				}
			}
		}

		explanation := explainedChanges(deleted, closestAdding)
		if !moveOnly && explanation != "" {
			method := method{
				before:       deleted,
				after:        closestAdding,
				commonFactor: commonFactor,
				explanation:  explanation,
			}
			methods = append(methods, method)
		}
	}

	return &methods, nil
}

// files initializes files struct
func files(changedFiles []string, b Break) ([]file, []file) {
	supported := make([]file, 0)
	ignored := make([]file, 0)

	for _, fileLine := range changedFiles {
		f := file{}
		status, name, filetype := extractDataFile(fileLine)
		f.name = name
		f.status = status
		f.typeFile = filetype
		diff, err := f.getDiff(b.startPoint, b.endPoint)
		if err == nil {
			f.diff = *diff
		}

		if f.canHaveBreak() {
			if f.isTypeSupported() {
				supported = append(supported, f)
			} else {
				ignored = append(ignored, f)
			}
		}
	}

	return supported, ignored
}

func (f *file) canHaveBreak() bool {
	return "A" != f.status
}

// explainedChanges try to understand nature of changes, returning a reason
// for compatibility break
func explainedChanges(before string, after string) string {
	if after == "" {
		return "Deletion of method"
	}

	deleted, added := differences(strings.Split(before, ","), strings.Split(after, ","))
	if len(deleted) > len(added) {
		return "Deletion of parameter"
	} else if len(deleted) < len(added) {
		var explanation string
		for i := 0; i < len(added); i++ {
			if strings.Contains(added[i], "=") {
				explanation = ""
				continue
			}
			explanation = "Adding a parameter without default value"
		}
		return explanation
	} else {
		explanation := "Unknown signature change"
		for i := 0; i < len(deleted); i++ {
			if strings.Contains(deleted[i], "=") && !strings.Contains(added[i], "=") {
				explanation = "Deletion of default parameter"
				break
			} else if !strings.Contains(added[i], "=") {
				explanation = "Adding a parameter without default value"
				break
			}
			// TODO : Precise cases :
			//	- add type
			//	- change type
			// 	- drop type (not a CB)
		}
		return explanation
	}
}

// differences shows a slice of differences between two slices
func differences(before []string, after []string) ([]string, []string) {
	var length int
	lengthBefore := len(before)
	lengthAfter := len(after)

	if lengthBefore < lengthAfter {
		for i := lengthBefore; i < lengthAfter; i++ {
			before = append(before, "")
		}
		length = lengthAfter
	} else {
		for i := lengthAfter; i < lengthBefore; i++ {
			after = append(after, "")
		}
		length = lengthBefore
	}

	var deleted []string
	var added []string
	for i := 0; i < length; i++ {
		if before[i] == after[i] {
			continue
		}
		if before[i] == "" {
			added = append(added, after[i])
		} else if after[i] == "" {
			deleted = append(deleted, before[i])
		} else {
			deleted = append(deleted, before[i])
			added = append(added, after[i])
		}
	}

	return deleted, added
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

// diff represents the diff of a file, segregated with deletion and adding
type diff struct {
	deletions []string
	addings   []string
}

// getDiff fetches diff (in a git sense) and extracts changes occured
func (f *file) getDiff(startObject string, endObject string) (*diff, error) {
	if f.isDeleted() {
		return f.getDiffDeleted(startObject)
	}
	command, err := qexec.Run("git", "diff", "-U0", startObject+"..."+endObject, f.name)
	if err != nil {
		return nil, err
	}

	pattern, errPattern := f.breakPattern()
	if errPattern != nil {
		return nil, errPattern
	}

	var diffDeleted []string
	var diffAdded []string
	for _, line := range strings.Split(command, "\n") {
		if strings.HasPrefix(line, "-") {
			diffDeleted = append(diffDeleted, strings.TrimSpace(line[1:]))
		} else if strings.HasPrefix(line, "+") {
			diffAdded = append(diffAdded, strings.TrimSpace(line[1:]))
		}
	}

	return &diff{
		deletions: filteredByPattern(pattern, diffDeleted),
		addings:   filteredByPattern(pattern, diffAdded),
	}, nil
}

func (f *file) isDeleted() bool {
	return "D" == f.status
}

func (f *file) getDiffDeleted(startObject string) (*diff, error) {
	fileDeleted, err := qexec.Run("git", "show", startObject+":"+f.name)
	if err != nil {
		return nil, err
	}

	pattern, errPattern := f.breakPattern()
	if errPattern != nil {
		return nil, errPattern
	}
	var diffDeleted []string
	for _, line := range strings.Split(fileDeleted, "\n") {
		diffDeleted = append(diffDeleted, strings.TrimSpace(line))
	}

	return &diff{
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

func (f *file) isTypeSupported() bool {
	_, err := f.breakPattern()

	return err == nil
}

// breakPattern returns the regex of a potential compatibility break associated
// with type of the file
func (f *file) breakPattern() (*regexp.Regexp, error) {
	var pattern *regexp.Regexp
	switch f.typeFile {
	case "go":
		pattern = regexp.MustCompile(`^(\s)*func( \(.+)\)? [A-Z]{1}[A-Za-z]*\(`)
	case "php":
		pattern = regexp.MustCompile(`^(\s)*public( static)? function [_A-Za-z]+\(|^(\s)*function [_A-Za-z]+\(`)
	case "java":
		pattern = regexp.MustCompile(`^(\s)*public( static)?( .+)? [A-Za-z]+\(`)
	case "js":
		pattern = regexp.MustCompile(`^(\s)*function [A-Za-z]+\(|^(\s)*(var )?[A-Za-z._]+(\s)*=(\s)*function \(|(\s)*[A-Za-z._]+(\s)*:(\s)*function \(`)
	case "sh":
		pattern = regexp.MustCompile(`^(\s)*function [A-Za-z_]+\(`)
	}

	if pattern == nil {
		return pattern, errors.New("Unknown langage")
	}

	return pattern, nil
}
