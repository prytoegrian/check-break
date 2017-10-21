package check

import (
	"errors"
	"strings"

	"github.com/tbruyelle/git"
	"github.com/tbruyelle/qexec"
)

func refExists(point string) bool {
	exists, _ := git.RefExists(point)
	return exists
}

func diffFileList(startPoint string, endPoint string) ([]string, error) {
	gitFiles, err := qexec.Run("git", "diff", "--name-status", startPoint+"..."+endPoint)
	if err != nil {
		return make([]string, 0), err
	}
	if len(gitFiles) == 0 {
		return make([]string, 0), errors.New("No changed file between these two points")
	}
	return strings.Split(strings.TrimSpace(gitFiles), "\n"), nil
}

func diffFile(startPoint string, endPoint string, filename string) ([]string, error) {
	diff, err := qexec.Run("git", "diff", "-U0", startPoint+"..."+endPoint, filename)
	if err != nil {
		return make([]string, 0), err
	}

	return strings.Split(diff, "\n"), nil
}

func showFile(startPoint string, filename string) ([]string, error) {
	diff, err := qexec.Run("git", "show", startPoint+":"+filename)
	if err != nil {
		return make([]string, 0), err
	}

	return strings.Split(diff, "\n"), nil
}
