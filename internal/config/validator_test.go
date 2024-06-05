package config

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestDeviceValidator(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *Config
		GetBlockDevice func(name string) (*model.BlockDevice, error)
		ExpectedError  error
	}{
		{
			Name: "Valid Device",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {},
				},
			},
			GetBlockDevice: func(name string) (*model.BlockDevice, error) {
				return &model.BlockDevice{Name: name}, nil
			},
			ExpectedError: nil,
		},
		{
			Name: "Device With Unsupported File System",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/vdb": {},
				},
			},
			GetBlockDevice: func(name string) (*model.BlockDevice, error) {
				return nil, fmt.Errorf("🔴 %s: 'jfs' is not a supported file system", name)
			},
			ExpectedError: fmt.Errorf("🔴 /dev/vdb: 'jfs' is not a supported file system"),
		},
	}
	for _, subtest := range subtests {
		ds := service.NewMockDeviceService()
		if subtest.GetBlockDevice != nil {
			ds.StubGetBlockDevice = subtest.GetBlockDevice
		}

		dv := NewDeviceValidator(ds)
		err := dv.Validate(subtest.Config)
		utils.CheckError("dv.Validate()", t, subtest.ExpectedError, err)
	}
}

func TestMountPointValidator(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *Config
		ExpectedError error
	}{
		{
			Name: "Valid Mount Point",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/app",
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid Mount Point (Relative Path)",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						MountPoint: "relative-path/app",
					},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: relative-path/app is not an absolute path"),
		},
		{
			Name: "Invalid Mount Point (Root Directory)",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						MountPoint: "/",
					},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Can not be mounted to the root directory"),
		},
	}
	for _, subtest := range subtests {
		mpv := NewMountPointValidator()
		err := mpv.Validate(subtest.Config)
		utils.CheckError("mpv.Validate()", t, subtest.ExpectedError, err)
	}
}

func TestMountOptionsValidator(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *Config
		ExpectedError error
	}{
		{
			Name: "Valid Mount Options",
			Config: &Config{
				Defaults: Options{
					MountOptions: "defaults",
				},
				Devices: map[string]Device{
					"/dev/xvdf": {
						Options: Options{
							MountOptions: "nouuid",
						},
					},
				},
				overrides: Options{
					MountOptions: "has_journal",
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid Mount Options (Bind, Overrides)",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {},
				},
				overrides: Options{
					MountOptions: "bind",
				},
			},
			ExpectedError: fmt.Errorf("🔴 'bind' (-mount-options) is not a supported mode as bind mounts are not supported for block devices"),
		},
		{
			Name: "Invalid Mount Options (Remount, Defaults)",
			Config: &Config{
				Defaults: Options{
					MountOptions: "remount",
				},
				Devices: map[string]Device{
					"/dev/xvdf": {},
				},
			},
			ExpectedError: fmt.Errorf("🔴 'remount' (defaults) is not a supported mode as it prevents unmounted devices from being mounted"),
		},
		{
			Name: "Invalid Mount Options (Bind, Device)",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						Options: Options{
							MountOptions: "bind",
						},
					},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: 'bind' is not a supported mode as bind mounts are not supported for block devices"),
		},
	}
	for _, subtest := range subtests {
		mov := NewMountOptionsValidator()
		err := mov.Validate(subtest.Config)
		utils.CheckError("mov.Validate()", t, subtest.ExpectedError, err)
	}
}

func TestFileSystemValidator(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *Config
		ExpectedError error
	}{
		{
			Name: "Valid File System",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
					},
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Unsupported File System",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						Fs: model.FileSystem("jfs"),
					},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: File system 'jfs' is not supported"),
		},
		{
			Name: "No File System Provided",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Must provide a supported file system"),
		},
		{
			Name: "LVM File System",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						Fs: model.Lvm,
					},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Refer to %s on how to manage LVM file systems", LvmWikiDocumentationUrl),
		},
	}
	for _, subtest := range subtests {
		fsv := NewFileSystemValidator()
		err := fsv.Validate(subtest.Config)
		utils.CheckError("fsv.Validate()", t, subtest.ExpectedError, err)
	}
}

const (
	Invalid = model.Mode("invalid")
)

func TestModeValidator(t *testing.T) {

	subtests := []struct {
		Name          string
		Config        *Config
		ExpectedError error
	}{
		{
			Name: "Valid Modes",
			Config: &Config{
				Defaults: Options{
					Mode: model.Prompt,
				},
				Devices: map[string]Device{
					"/dev/xvdf": {
						Options: Options{
							Mode: model.Healthcheck,
						},
					},
				},
				overrides: Options{
					Mode: model.Force,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid Mode (Overrides)",
			Config: &Config{
				Defaults: Options{
					Mode: model.Force,
				},
				Devices: map[string]Device{
					"/dev/xvdf": {
						Options: Options{
							Mode: model.Prompt,
						},
					},
				},
				overrides: Options{
					Mode: Invalid,
				},
			},
			ExpectedError: fmt.Errorf("🔴 '%s' (-mode) is not a supported mode", Invalid),
		},
		{
			Name: "Invalid Mode (Defaults)",
			Config: &Config{
				Defaults: Options{
					Mode: Invalid,
				},
				Devices: map[string]Device{
					"/dev/xvdf": {
						Options: Options{
							Mode: model.Prompt,
						},
					},
				},
			},
			ExpectedError: fmt.Errorf("🔴 '%s' (defaults) is not a supported mode", Invalid),
		},
		{
			Name: "Invalid Mode (Device)",
			Config: &Config{
				Defaults: Options{
					Mode: model.Force,
				},
				Devices: map[string]Device{
					"/dev/xvdf": {
						Options: Options{
							Mode: Invalid,
						},
					},
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: '%s' is not a supported mode", Invalid),
		},
	}
	for _, subtest := range subtests {
		mv := NewModeValidator()
		err := mv.Validate(subtest.Config)
		utils.CheckError("mv.Validate()", t, subtest.ExpectedError, err)
	}
}

func TestOwnerValidator(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *Config
		GetUser       func(usr string) (*model.User, error)
		GetGroup      func(grp string) (*model.Group, error)
		ExpectedError error
	}{
		{
			Name: "Valid User and Valid Group",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						User:  "example",
						Group: "example",
					},
				},
			},
			GetUser: func(usr string) (*model.User, error) {
				return &model.User{
					Name: usr,
					Id:   1000,
				}, nil
			},
			GetGroup: func(grp string) (*model.Group, error) {
				return &model.Group{
					Name: grp,
					Id:   2000,
				}, nil
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid User and Valid Group",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						User:  "example",
						Group: "example",
					},
				},
			},
			GetUser: func(usr string) (*model.User, error) {
				return nil, fmt.Errorf("🔴 User (name=%s) does not exist", usr)
			},
			GetGroup: func(grp string) (*model.Group, error) {
				return &model.Group{
					Name: grp,
					Id:   2000,
				}, nil
			},
			ExpectedError: fmt.Errorf("🔴 User (name=example) does not exist"),
		},
		{
			Name: "Valid User and Invalid Group",
			Config: &Config{
				Devices: map[string]Device{
					"/dev/xvdf": {
						User:  "example",
						Group: "example",
					},
				},
			},
			GetUser: func(usr string) (*model.User, error) {
				return &model.User{
					Name: usr,
					Id:   1000,
				}, nil
			},
			GetGroup: func(grp string) (*model.Group, error) {
				return nil, fmt.Errorf("🔴 Group (name=%s) does not exist", grp)
			},
			ExpectedError: fmt.Errorf("🔴 Group (name=example) does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			owns := service.NewMockOwnerService()
			if subtest.GetUser != nil {
				owns.StubGetUser = subtest.GetUser
			}
			if subtest.GetGroup != nil {
				owns.StubGetGroup = subtest.GetGroup
			}

			ov := NewOwnerValidator(owns)
			err := ov.Validate(subtest.Config)
			utils.CheckError("ov.Validate()", t, subtest.ExpectedError, err)
		})
	}
}
