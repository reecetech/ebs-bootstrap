package service

import (
	"fmt"
	"os"
	"testing"
	"ebs-bootstrap/internal/utils"
	"github.com/google/go-cmp/cmp"
)

func TestGetStats(t *testing.T) {
	fs := &UnixFileService{}
	t.Run("Get File Stats (Existing File)", func(t *testing.T) {
		owner, group, permissions := os.Getuid(), os.Getgid(), os.FileMode(0644)
		f, err := os.CreateTemp("", "sample")
		utils.CheckError("CreateTemp()", t, nil, err)
		defer os.Remove(f.Name())

		err = os.Chown(f.Name(), owner, group)
		utils.CheckError("Chown()", t, nil, err)

		err = os.Chmod(f.Name(), permissions)
		utils.CheckError("Chmod()", t, nil, err)

		actual, err := fs.GetStats(f.Name())
		utils.CheckError("GetStats()", t, nil, err)
		
		expected := &FileInfo{
			Owner: fmt.Sprintf("%d", owner),
			Group: fmt.Sprintf("%d", group),
			Permissions: fmt.Sprintf("%o", permissions),
			Exists: true,
		}
		if !cmp.Equal(actual, expected) {
			t.Errorf("GetStats() [output] mismatch: Expected=%+v Actual=%+v", expected, actual)
		}
	})
	t.Run("Get File Stats (Non-Existent File)", func(t *testing.T) {
		expected := &FileInfo{Exists: false}
		actual, err := fs.GetStats("/non-existent-file/file.txt")
		if !cmp.Equal(actual, expected) {
			t.Errorf("GetStats() [output] mismatch: Expected=%+v Actual=%+v", expected, actual)
		}
		utils.CheckError("GetStats()", t, nil, err)
	})
}

func TestValidateFile(t *testing.T) {
	fs := &UnixFileService{}

	// Create a variable to the current working directory
	d, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() [error] %s", err)
		return
	}

	// Create a temporary file
	f, err := os.CreateTemp("", "validate-file")
	if err != nil {
		t.Errorf("os.CreateTemp() [error] %s", err)
		return
	}
	defer os.Remove(f.Name())

	subtests := []struct{
		Name		string
		Path		string
		ExpectedErr	error
	}{
		{
			Name: 			"Valid (Existing File)",
			Path:			f.Name(),
			ExpectedErr:	nil,
		},
		{
			Name: 			"Invalid (Existing Directory)",
			Path:			d,
			ExpectedErr:	fmt.Errorf("🔴 %s is not a regular file", d),
		},
		{
			Name: 			"Invalid: (Non-existing File)",
			Path:			"/doesnt-exist",
			ExpectedErr:	fmt.Errorf("🔴 /doesnt-exist does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			err := fs.ValidateFile(subtest.Path)
			utils.CheckError("ValidateFile()", t, subtest.ExpectedErr, err)
		})
	}
}

func TestValidateDirectory(t *testing.T) {
	fs := &UnixFileService{}

	// Create a variable to the current working directory
	d, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() [error] %s", err)
		return
	}

	// Create a temporary file
	f, err := os.CreateTemp("", "validate-directory")
	if err != nil {
		t.Errorf("os.CreateTemp() [error] %s", err)
		return
	}
	defer os.Remove(f.Name())

	subtests := []struct{
		Name		string
		Path		string
		ExpectedErr	error
	}{
		{
			Name: 			"Valid (Existing Directory)",
			Path:			d,
			ExpectedErr:	nil,
		},
		{
			Name: 			"Invalid (Existing File)",
			Path:			f.Name(),
			ExpectedErr:	fmt.Errorf("🔴 %s is not a directory", f.Name()),
		},
		{
			Name: 			"Invalid (Non-existing Directory)",
			Path:			"/doesnt-exist",
			ExpectedErr:	fmt.Errorf("🔴 /doesnt-exist does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			err := fs.ValidateDirectory(subtest.Path)
			utils.CheckError("ValidateDirectory()", t, subtest.ExpectedErr, err)
		})
	}
}
