package docker

import (
	"github.com/dimuls/mycode"
)

const (
	cSrcPath     = "/c"
	cSrcFileName = "main.c"
)

var (
	cCompileCmd = []string{"gcc", cSrcFileName}
	cRunCmd     = []string{"./a.out"}
)

type CRunPreparator struct{}

func (p CRunPreparator) Prepare(source string) (Run, error) {

	err := makeSrcPath(cSrcPath, cSrcFileName, source)
	if err != nil {
		return nil, err
	}

	return simpleRun{
		sourcePath:     cSrcPath,
		compileCommand: cCompileCmd,
		runCommand:     cRunCmd,
	}, nil
}

func init() {
	RegisterRunPreparator(mycode.Language_c, CRunPreparator{})
}
