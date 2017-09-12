package check

import (
	"fmt"

	"github.com/fatih/color"
)

// BreakReport is a report to display
type BreakReport struct {
	Supported  []FileReport
	Ignored    []file
	Exclusions []string
}

// Report displays a BreakReport
func (b *Break) Report() (*BreakReport, error) {
	f, err := diffFileList(b.startPoint, b.endPoint)
	if err != nil {
		return nil, err
	}
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

func (f *file) Report() string {
	return fmt.Sprint(">> ", color.CyanString(f.name))
}
