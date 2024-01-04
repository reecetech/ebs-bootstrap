package model

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/utils"
	"gopkg.in/yaml.v2"
)

func TestUnmarshalYAML(t *testing.T) {
	subtests := []struct {
		Name           string
		Yaml           []byte
		ExpectedOutput FilePermissions
		ExpectedError  error
	}{
		{
			Name:           "Valid + Empty",
			Yaml:           []byte(`""`),
			ExpectedOutput: FilePermissions(0),
			ExpectedError:  nil,
		},
		{
			Name:           "Valid",
			Yaml:           []byte("755"),
			ExpectedOutput: FilePermissions(0755),
			ExpectedError:  nil,
		},
		{
			Name:           "Valid + Octal",
			Yaml:           []byte("0755"),
			ExpectedOutput: FilePermissions(0755),
			ExpectedError:  nil,
		},
		{
			Name:           "Invalid + Incorrect Base",
			Yaml:           []byte("0892"),
			ExpectedOutput: FilePermissions(0),
			ExpectedError:  fmt.Errorf("🔴 invalid permission value. '0892' must be a valid octal number"),
		},
		{
			Name:           "Invalid + Exceeds Maximum File Permissions (0777)",
			Yaml:           []byte("1777"),
			ExpectedOutput: FilePermissions(0),
			ExpectedError:  fmt.Errorf("🔴 invalid permission value. '01777' exceeds the maximum allowed value (0777)"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			var fp FilePermissions
			err := yaml.Unmarshal(subtest.Yaml, &fp)
			utils.CheckError("yaml.Unmarshal()", t, subtest.ExpectedError, err)
			utils.CheckOutput("yaml.Unmarshal()", t, subtest.ExpectedOutput, fp)
		})
	}
}

func TestPerm(t *testing.T) {
	subtests := []struct {
		Name           string
		FilePermission FilePermissions
		ExpectedOutput fs.FileMode
	}{
		{
			Name:           "Valid",
			FilePermission: FilePermissions(0755),
			ExpectedOutput: fs.FileMode(0755),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			utils.CheckOutput("FilePermission.Perm()", t, subtest.ExpectedOutput, subtest.FilePermission.Perm())
		})
	}
}
