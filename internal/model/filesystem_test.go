package model

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestParseFileSystem(t *testing.T) {
	subtests := []struct {
		Name           string
		FileSystem     string
		ExpectedOutput FileSystem
		ExpectedError  error
	}{
		{
			FileSystem:     "",
			ExpectedOutput: Unformatted,
			ExpectedError:  nil,
		},
		{
			FileSystem:     "xfs",
			ExpectedOutput: Xfs,
			ExpectedError:  nil,
		},
		{
			FileSystem:     "ext4",
			ExpectedOutput: Ext4,
			ExpectedError:  nil,
		},
		{
			FileSystem:     "jfs",
			ExpectedOutput: FileSystem("jfs"),
			ExpectedError:  fmt.Errorf("File system 'jfs' is not supported"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.FileSystem, func(t *testing.T) {
			fs, err := ParseFileSystem(subtest.FileSystem)
			utils.CheckError("ParseFileSystem()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ParseFileSystem()", t, subtest.ExpectedOutput, fs)
		})
	}
}

func TestFileSystemString(t *testing.T) {
	subtests := []struct {
		Name           string
		FileSystem     FileSystem
		ExpectedOutput string
	}{
		{
			FileSystem:     Unformatted,
			ExpectedOutput: "unformatted",
		},
		{
			FileSystem:     Xfs,
			ExpectedOutput: "xfs",
		},
		{
			FileSystem:     Ext4,
			ExpectedOutput: "ext4",
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.FileSystem.String(), func(t *testing.T) {
			utils.CheckOutput("FileSystem.String()", t, subtest.ExpectedOutput, subtest.FileSystem.String())
		})
	}
}
