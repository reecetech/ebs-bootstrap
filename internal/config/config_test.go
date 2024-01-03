package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestParsing(t *testing.T) {
	subtests := []struct {
		Name           string
		Data           []byte
		ExpectedOutput *Config
		ExpectedError  error
	}{
		{
			Name: "Valid Config",
			Data: []byte(`---
defaults:
  mode: healthcheck
devices:
  /dev/xvdf:
    fs: xfs
    mountPoint: /ifmx/dev/root
    user: 0
    group: root
    permissions: 755
    label: external-vol
    resizeFs: true
    resizeThreshold: 95
    remount: true`),
			ExpectedOutput: &Config{
				Defaults: Options{
					Mode: model.Healthcheck,
				},
				Devices: map[string]Device{
					"/dev/xvdf": {
						Fs:          model.Xfs,
						MountPoint:  "/ifmx/dev/root",
						User:        "0",
						Group:       "root",
						Permissions: model.FilePermissions(0755),
						Label:       "external-vol",
						Options: Options{
							ResizeFs:        true,
							ResizeThreshold: 95,
							Remount:         true,
						},
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name:           "Unsupported Attribute",
			Data:           []byte(`unsupported: true`),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /tmp/*.yml: Failed to ingest malformed config"),
		},
		{
			Name:           "Malformed YAML",
			Data:           []byte(`malformed:- true`),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /tmp/*.yml: Failed to ingest malformed config"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			configPath, err := createConfigFile(subtest.Data)
			utils.CheckError("createConfigFile()", t, nil, err)
			defer os.Remove(configPath)

			c, err := New([]string{"ebs-bootstrap", "-config", configPath})
			utils.CheckErrorGlob("config.New()", t, subtest.ExpectedError, err)
			// Config contains the unexported attribute "overrides"
			// We need to allow go-cmp to inspect the contents of unexported attributes
			utils.CheckOutput("config.New()", t, subtest.ExpectedOutput, c, cmp.AllowUnexported(Config{}))
		})
	}
}

func TestFlagParsing(t *testing.T) {
	// Create a variable to the current working directory
	d, err := os.Getwd()
	utils.CheckError("os.Getwd()", t, nil, err)

	subtests := []struct {
		Name          string
		Args          []string
		ExpectedError error
	}{
		{
			Name:          "Invalid Config (Directory)",
			Args:          []string{"ebs-bootstrap", "-config", d},
			ExpectedError: fmt.Errorf("🔴 %s: *", d),
		},
		{
			Name:          "Invalid Config (Non-existent File)",
			Args:          []string{"ebs-bootstrap", "-config", "/doesnt-exist"},
			ExpectedError: fmt.Errorf("🔴 /doesnt-exist: File not found"),
		},
		{
			Name:          "Unsupported Flag",
			Args:          []string{"ebs-bootstrap", "-unsupported-flag"},
			ExpectedError: fmt.Errorf("🔴 Failed to parse provided flags"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			_, err := New(subtest.Args)
			utils.CheckErrorGlob("config.New()", t, subtest.ExpectedError, err)
		})
	}
}

func TestOptions(t *testing.T) {
	device := "/dev/xvdf"
	subtests := []struct {
		Name           string
		Data           []byte
		ExpectedOutput *Options
		ExpectedError  error
	}{
		{
			Name: "Provide Non-Default Device Options",
			Data: []byte(fmt.Sprintf(`---
devices:
  %s:
    mode: prompt
    remount: true
    mountOptions: nouuid
    resizeFs: true
    resizeThreshold: 95`, device)),
			ExpectedOutput: &Options{
				Mode:            model.Prompt,
				Remount:         true,
				MountOptions:    "nouuid",
				ResizeFs:        true,
				ResizeThreshold: 95,
			},
			ExpectedError: nil,
		},
		{
			Name: "Default Options for Non-Existent Device",
			Data: []byte(`---
devices:
  /dev/nonexist: ~`),
			ExpectedOutput: &Options{
				Mode:            model.Healthcheck,
				Remount:         false,
				MountOptions:    "defaults",
				ResizeFs:        false,
				ResizeThreshold: 0,
			},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			configPath, err := createConfigFile(subtest.Data)
			utils.CheckError("createConfigFile()", t, nil, err)
			defer os.Remove(configPath)

			c, err := New([]string{"ebs-bootstrap", "-config", configPath})
			utils.CheckError("config.New()", t, subtest.ExpectedError, err)

			d := &Options{
				Mode:            c.GetMode(device),
				Remount:         c.GetRemount(device),
				MountOptions:    c.GetMountOptions(device),
				ResizeFs:        c.GetResizeFs(device),
				ResizeThreshold: c.GetResizeThreshold(device),
			}
			utils.CheckOutput("config.New()", t, subtest.ExpectedOutput, d)
		})
	}
}

func TestFlagOptions(t *testing.T) {
	device := "/dev/xvdf"
	c, err := createConfigFile([]byte(fmt.Sprintf(`---
devices:
  %s: ~`, device)))
	utils.CheckError("createConfigFile()", t, nil, err)
	defer os.Remove(c)
	subtests := []struct {
		Name           string
		Args           []string
		ExpectedOutput *Options
		ExpectedError  error
	}{
		{
			Name: "Mode Flag Options",
			Args: []string{"ebs-bootstrap", "-config", c, "-mode", string(model.Force)},
			ExpectedOutput: &Options{
				Mode:            model.Force,
				Remount:         false,
				MountOptions:    "defaults",
				ResizeFs:        false,
				ResizeThreshold: 0,
			},
			ExpectedError: nil,
		},
		{
			Name: "Mount Flag Options",
			Args: []string{"ebs-bootstrap", "-config", c, "-remount", "-mount-options", "nouuid"},
			ExpectedOutput: &Options{
				Mode:            model.Healthcheck,
				Remount:         true,
				MountOptions:    "nouuid",
				ResizeFs:        false,
				ResizeThreshold: 0,
			},
			ExpectedError: nil,
		},
		{
			Name: "Resize Flag Options",
			Args: []string{"ebs-bootstrap", "-config", c, "-resize-fs", "-resize-threshold", "95"},
			ExpectedOutput: &Options{
				Mode:            model.Healthcheck,
				Remount:         false,
				MountOptions:    "defaults",
				ResizeFs:        true,
				ResizeThreshold: 95,
			},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			c, err := New(subtest.Args)
			utils.CheckError("config.New()", t, subtest.ExpectedError, err)

			o := &Options{
				Mode:            c.GetMode(device),
				Remount:         c.GetRemount(device),
				MountOptions:    c.GetMountOptions(device),
				ResizeFs:        c.GetResizeFs(device),
				ResizeThreshold: c.GetResizeThreshold(device),
			}
			utils.CheckOutput("config.New()", t, subtest.ExpectedOutput, o)
		})
	}
}

func TestDefaultOptions(t *testing.T) {
	device := "/dev/xvdf"
	subtests := []struct {
		Name           string
		Data           []byte
		ExpectedOutput *Options
		ExpectedError  error
	}{
		{
			Name: "Mode Default Options",
			Data: []byte(fmt.Sprintf(`---
defaults:
  mode: force
devices:
  %s: ~`, device)),
			ExpectedOutput: &Options{
				Mode:            model.Force,
				Remount:         false,
				MountOptions:    "defaults",
				ResizeFs:        false,
				ResizeThreshold: 0,
			},
			ExpectedError: nil,
		},
		{
			Name: "Mount Default Options",
			Data: []byte(fmt.Sprintf(`---
defaults:
  remount: true
  mountOptions: nouuid
devices:
  %s: ~`, device)),
			ExpectedOutput: &Options{
				Mode:            model.Healthcheck,
				Remount:         true,
				MountOptions:    "nouuid",
				ResizeFs:        false,
				ResizeThreshold: 0,
			},
			ExpectedError: nil,
		},
		{
			Name: "Resize Default Options",
			Data: []byte(fmt.Sprintf(`---
defaults:
  resizeFs: true
  resizeThreshold: 95
devices:
  %s: ~`, device)),
			ExpectedOutput: &Options{
				Mode:            model.Healthcheck,
				Remount:         false,
				MountOptions:    "defaults",
				ResizeFs:        true,
				ResizeThreshold: 95,
			},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			configPath, err := createConfigFile(subtest.Data)
			utils.CheckError("createConfigFile()", t, nil, err)
			defer os.Remove(configPath)

			c, err := New([]string{"ebs-bootstrap", "-config", configPath})
			utils.CheckError("config.New()", t, subtest.ExpectedError, err)

			d := &Options{
				Mode:            c.GetMode(device),
				Remount:         c.GetRemount(device),
				MountOptions:    c.GetMountOptions(device),
				ResizeFs:        c.GetResizeFs(device),
				ResizeThreshold: c.GetResizeThreshold(device),
			}
			utils.CheckOutput("config.New()", t, subtest.ExpectedOutput, d)
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
