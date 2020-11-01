package docker

import "github.com/dimuls/mycode"

const (
	goSrcPath     = "/go/src/app"
	goSrcFileName = "main.go"
)

var (
	goCompileCmd = []string{"go", "build", goSrcFileName}
	goRunCmd     = []string{"./main"}
)

type GoRunPreparator struct{}

func (p GoRunPreparator) Prepare(source string) (Run, error) {

	err := makeSrcPath(goSrcPath, goSrcFileName, source)
	if err != nil {
		return nil, err
	}

	return simpleRun{
		sourcePath:     goSrcPath,
		compileCommand: goCompileCmd,
		runCommand:     goRunCmd,
	}, nil
}

func init() {
	RegisterRunPreparator(mycode.Language_go, GoRunPreparator{})
}
