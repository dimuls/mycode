package docker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/dimuls/mycode"
)

type Run interface {
	SourcePath() (string, error)
	AdditionalBinds() ([]string, error)
	CompileCommand() ([]string, error)
	RunCommand() ([]string, error)
}

type RunPreparator interface {
	Prepare(source string) (Run, error)
}

var preparators = map[mycode.Language]RunPreparator{}

func RegisterRunPreparator(language mycode.Language, cp RunPreparator) {
	preparators[language] = cp
}

func Prepare(language mycode.Language, source string) (
	Run, error) {

	p, exists := preparators[language]
	if !exists {
		return nil, fmt.Errorf("unknown language")
	}

	r, err := p.Prepare(source)
	if err != nil {
		return nil, fmt.Errorf("prepare code: %w", err)
	}

	return r, nil
}

func makeSrcPath(srcPath, srcFileName, src string) error {
	err := os.MkdirAll(srcPath, 0755)
	if err != nil {
		return fmt.Errorf("make soruce path directory: %w", err)
	}

	err = ioutil.WriteFile(filepath.Join(srcPath, srcFileName),
		[]byte(src), 0644)
	if err != nil {
		return fmt.Errorf("write source file: %w", err)
	}

	return nil
}

type simpleRun struct {
	sourcePath      string
	additionalBinds []string
	compileCommand  []string
	runCommand      []string
}

func (sr simpleRun) SourcePath() (string, error) {
	return sr.sourcePath, nil
}

func (sr simpleRun) AdditionalBinds() ([]string, error) {
	return sr.additionalBinds, nil
}

func (sr simpleRun) CompileCommand() ([]string, error) {
	return sr.compileCommand, nil
}

func (sr simpleRun) RunCommand() ([]string, error) {
	return sr.runCommand, nil
}
