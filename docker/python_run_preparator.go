package docker

import "github.com/dimuls/mycode"

const (
	pythonSrcPath     = "/python"
	pythonSrcFileName = "main.py"
)

var pythonRunCmd = []string{"python3", pythonSrcFileName}

type PythonRunPreparator struct{}

func (p PythonRunPreparator) Prepare(source string) (Run, error) {

	err := makeSrcPath(pythonSrcPath, pythonSrcFileName, source)
	if err != nil {
		return nil, err
	}

	return simpleRun{
		sourcePath:      pythonSrcPath,
		additionalBinds: []string{"/usr/bin/python3"},
		runCommand:      pythonRunCmd,
	}, nil
}

func init() {
	RegisterRunPreparator(mycode.Language_python, PythonRunPreparator{})
}
