package docker

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/dimuls/mycode"
)

const (
	javaSrcPath     = "/java"
	javaSrcFileName = "main.java"
	classSuffix     = ".class"
)

var (
	javaCompileCmd = []string{"/usr/lib/jvm/java-11-openjdk/bin/javac", javaSrcFileName}
	javaRunCmd     = []string{"/usr/lib/jvm/java-11-openjdk/bin/java"}
)

type javaRun struct {
	class string
}

func (jr javaRun) SourcePath() (string, error) {
	return javaSrcPath, nil
}

func (jr javaRun) AdditionalBinds() ([]string, error) {
	return nil, nil
}

func (jr javaRun) CompileCommand() ([]string, error) {
	return javaCompileCmd, nil
}

func (jr javaRun) RunCommand() ([]string, error) {
	dir, err := ioutil.ReadDir(javaSrcPath)
	if err != nil {
		return nil, fmt.Errorf("read source dir: %w", err)
	}

	for _, f := range dir {
		if f.IsDir() {
			continue
		}
		if strings.HasSuffix(f.Name(), classSuffix) {
			className := strings.TrimRight(filepath.Base(f.Name()), classSuffix)
			return append(javaRunCmd, className), nil
		}
	}

	return nil, fmt.Errorf("no class file found")
}

type JavaRunPreparator struct{}

func (p JavaRunPreparator) Prepare(source string) (Run, error) {

	err := makeSrcPath(javaSrcPath, javaSrcFileName, source)
	if err != nil {
		return nil, err
	}

	return javaRun{}, nil
}

func init() {
	RegisterRunPreparator(mycode.Language_java, JavaRunPreparator{})
}
