package utils

import (
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

type Binary string

const (
	Lsblk     Binary = "lsblk"
	MkfsExt4  Binary = "mkfs.ext4"
	E2Label   Binary = "e2label"
	MkfsXfs   Binary = "mkfs.xfs"
	XfsAdmin  Binary = "xfs_admin"
	Mount     Binary = "mount"
	Umount    Binary = "umount"
	BlockDev  Binary = "blockdev"
	Tune2fs   Binary = "tune2fs"
	XfsInfo   Binary = "xfs_info"
	Resize2fs Binary = "resize2fs"
	XfsGrowfs Binary = "xfs_growfs"
	PvCreate  Binary = "pvcreate"
)

type RunnerFactory interface {
	Select(binary Binary) Runner
}

type MockRunnerFactory struct {
	expectedBinary Binary
	expectedArgs   []string
	output         string
	err            error
}

func NewMockRunnerFactory(expectedBinary Binary, expectedArgs []string, output string, err error) *MockRunnerFactory {
	return &MockRunnerFactory{
		expectedBinary: expectedBinary,
		expectedArgs:   expectedArgs,
		output:         output,
		err:            err,
	}
}

func (mrf *MockRunnerFactory) Select(binary Binary) Runner {
	if mrf.expectedBinary != binary {
		return NewMockRunner(mrf.expectedArgs, "", fmt.Errorf("🔴 Unexpected Binary encountered: Expected=%s, Actual=%s", mrf.expectedBinary, binary))
	}
	return NewMockRunner(mrf.expectedArgs, mrf.output, mrf.err)
}

type ExecRunnerFactory struct {
	runners map[Binary]*ExecRunner
}

func NewExecRunnerFactory() *ExecRunnerFactory {
	return &ExecRunnerFactory{
		runners: map[Binary]*ExecRunner{},
	}
}

// Caching behaviour is implemented for ExecRunnerFactory as we
// do not need to validate an ExecRunner more than once
func (rc *ExecRunnerFactory) Select(binary Binary) Runner {
	r, exists := rc.runners[binary]
	if !exists {
		r = NewExecRunner(binary)
		rc.runners[binary] = r
	}
	return r
}

type Runner interface {
	Command(arg ...string) (string, error)
}

type MockRunner struct {
	expectedArgs []string
	output       string
	err          error
}

func NewMockRunner(expectedArgs []string, output string, err error) *MockRunner {
	return &MockRunner{
		expectedArgs: expectedArgs,
		output:       output,
		err:          err,
	}
}

func (mr *MockRunner) Command(arg ...string) (string, error) {
	if !slices.Equal(mr.expectedArgs, arg) {
		return "", fmt.Errorf("🔴 Unexpected arguments encountered. Expected=%v, Actual=%v", mr.expectedArgs, arg)
	}
	return mr.output, mr.err
}

type ExecRunner struct {
	binary      Binary
	command     func(name string, arg ...string) *exec.Cmd
	lookPath    func(file string) (string, error)
	isValidated bool
}

func NewExecRunner(binary Binary) *ExecRunner {
	return &ExecRunner{
		binary:   binary,
		command:  exec.Command,
		lookPath: exec.LookPath,
	}
}

func (er *ExecRunner) Command(arg ...string) (string, error) {
	if !er.isValid() {
		return "", fmt.Errorf("🔴 %s is either not installed or accessible from $PATH", string(er.binary))
	}
	cmd := er.command(string(er.binary), arg...)
	o, err := cmd.CombinedOutput()
	output := strings.TrimRight(string(o), "\n")
	if err != nil {
		return "", fmt.Errorf("🔴 %s: %s", err, output)
	}
	return output, err
}

func (er *ExecRunner) isValid() bool {
	if !er.isValidated {
		_, err := er.lookPath(string(er.binary))
		er.isValidated = err == nil
	}
	return er.isValidated
}
