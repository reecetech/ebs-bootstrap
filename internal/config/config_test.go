package config

import (
	"fmt"
	"os"
	"testing"
	"ebs-bootstrap/internal/utils"
	"ebs-bootstrap/internal/service"
	"github.com/google/go-cmp/cmp"
)

var dt = &service.DeviceTranslator{
	Table:	map[string]string{
		"/dev/xvdf": 	"/dev/nvme0n1",
		"/dev/nvme0n1": "/dev/nvme0n1",
	},
}

var fs = &service.UnixFileService{}

func TestConfigParsing(t *testing.T) {
	u, g, err := utils.GetCurrentUserGroup()
	if err != nil {
		t.Error(err)
		return
	}
	subtests := []struct{
		Name				string
		Data				[]byte
		ExpectedOutput		*Config
		ExpectedErr			error
	}{
        {
            Name: "Valid Config",
            Data: []byte(fmt.Sprintf(`---
global:
  mode: healthcheck
devices:
  /dev/xvdf:
    fs: "xfs"
    mount_point: "/ifmx/dev/root"
    owner: "%s"
    group: "%s"
    permissions: 755
    label: "external-vol"`,u.Name, g.Name)),
			ExpectedOutput: &Config{
                Global: ConfigGlobal{
                    Mode: "healthcheck",
                },
                Devices: map[string]ConfigDevice{
                    "/dev/nvme0n1": ConfigDevice{
                        Fs:   "xfs",
                        MountPoint: "/ifmx/dev/root",
                        Owner:  u.Uid,
                        Group:  g.Gid,
                        Permissions: "755",
                        Label:  "external-vol",
                        Mode: "healthcheck",
                    },
                },
            },
            ExpectedErr: nil,
        },
        {
            Name: "Malformed Config",
            Data: []byte(`---
global:
  mode: healthcheck
devices::
  /dev/xvdf:
    bad_attribute: false`),
			ExpectedOutput: nil,
            ExpectedErr: fmt.Errorf("🔴 Failed to ingest malformed config"),
        },
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			configPath, err := createConfigFile(subtest.Data)
			if err != nil {
				t.Errorf("createConfigFile() [error] %s", err)
			}
			defer os.Remove(configPath)

			c, err := New(
				[]string{"ebs-bootstrap-test", "-config", configPath},
				dt,
				fs,
			)
			if !cmp.Equal(c, subtest.ExpectedOutput) {
				t.Errorf("Modify() [output] mismatch: Expected=%+v Actual=%+v", subtest.ExpectedOutput, c)
			}
			utils.CheckError("config.New()", t, subtest.ExpectedErr, err)
		})
	}
}

func TestFlagParsing(t *testing.T) {
	u, g, err := utils.GetCurrentUserGroup()
	if err != nil {
		t.Error(err)
		return
	}
	// Create a variable to the path of a valid config
	c, err := createConfigFile([]byte(fmt.Sprintf(`---
global:
  mode: healthcheck
devices:
  /dev/xvdf:
    fs: "xfs"
    mount_point: "/ifmx/dev/root"
    owner: "%s"
    group: "%s"
    permissions: 755
    label: "external-vol"`,u.Uid, g.Gid)))
	if err != nil {
		t.Errorf("createConfigFile() [error] %s", err)
		return
	}

	// Create a variable to the current working directory
	d, err := os.Getwd()
	if err != nil {
		t.Errorf("os.Getwd() [error] %s", err)
	}

	subtests := []struct{
		Name		string
		Args		[]string
		ExpectedErr	error
	}{
		{
			Name: 			"Valid Config",
			Args:			[]string{"ebs-bootstrap-test","-config",c},
			ExpectedErr:	nil,
		},
		{
			Name: 			"Invalid Config (Directory)",
			Args:			[]string{"ebs-bootstrap-test","-config",d},
			ExpectedErr:	fmt.Errorf("🔴 %s is not a regular file", d),
		},
		{
			Name: 			"Invalid Config (Non-existent File)",
			Args:			[]string{"ebs-bootstrap-test","-config","/doesnt-exist"},
			ExpectedErr:	fmt.Errorf("🔴 /doesnt-exist does not exist"),
		},
		{
			Name: 			"Unsupported Flag",
			Args:			[]string{"ebs-bootstrap-test","-unsupported-flag"},
			ExpectedErr:	fmt.Errorf("🔴 Failed to parse provided flags"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			_, err := New(
				subtest.Args,
				dt,
				fs,
			)
			utils.CheckError("config.New()", t, subtest.ExpectedErr, err)
		})
	}
}

func createConfigFile(data []byte) (string, error) {
	f, err := os.CreateTemp("", "config_test_*.yml")
	if err != nil {
		return "", fmt.Errorf("🔴 Failed to create temporary config file: %v", err)
	}
	defer f.Close()
	if _, err := f.Write(data); err != nil {
		return "", fmt.Errorf("🔴 Failed to write to temporary config file: %v", err)
	}
	return f.Name(), nil
}
