package docker

import "github.com/dimuls/mycode"

const (
	cppSrcPath     = "/cpp"
	cppSrcFileName = "main.cpp"
)

var (
	cppCompileCmd = []string{"g++", cppSrcFileName}
	cppRunCmd     = []string{"./a.out"}
)

type CPPRunPreparator struct{}

func (p CPPRunPreparator) Prepare(source string) (Run, error) {

	err := makeSrcPath(cppSrcPath, cppSrcFileName, source)
	if err != nil {
		return nil, err
	}

	return simpleRun{
		sourcePath:     cppSrcPath,
		compileCommand: cppCompileCmd,
		runCommand:     cppRunCmd,
	}, nil
}

func init() {
	RegisterRunPreparator(mycode.Language_cpp, CPPRunPreparator{})
}
