package utils

import (
	"os/exec"
	"strings"
	"fmt"
)

type Runner interface {
	Command(name string, arg ...string) (string, error)
}

type ExecRunner struct {
	Runner func(name string, arg ...string) *exec.Cmd
}

func NewExecRunner() *ExecRunner {
	return &ExecRunner{exec.Command}
}

func (er *ExecRunner) Command(name string, arg ...string) (string, error) {
	cmd := er.Runner(name, arg...)
	o, err := cmd.CombinedOutput()
	output := strings.TrimRight(string(o), "\n")
	if err != nil {
		return "", fmt.Errorf("%s: %s", err, output)
	}
	return output, err
}
