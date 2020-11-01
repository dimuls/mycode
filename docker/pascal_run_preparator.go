package docker

import "github.com/dimuls/mycode"

const (
	pascalSrcPath     = "/pascal"
	pascalSrcFileName = "main.pas"
)

var (
	pascalCompileCmd = []string{"fpc", pascalSrcFileName}
	pascalRunCmd     = []string{"./main"}
)

type PascalRunPreparator struct{}

func (p PascalRunPreparator) Prepare(source string) (Run, error) {

	err := makeSrcPath(pascalSrcPath, pascalSrcFileName, source)
	if err != nil {
		return nil, err
	}

	return simpleRun{
		sourcePath:     pascalSrcPath,
		compileCommand: pascalCompileCmd,
		runCommand:     pascalRunCmd,
	}, nil
}

func init() {
	RegisterRunPreparator(mycode.Language_pascal, PascalRunPreparator{})
}
