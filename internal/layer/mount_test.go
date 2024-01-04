package layer

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestMountDeviceLayerModify(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Devices       map[string]*model.BlockDevice
		Files         map[string]*model.File
		CmpOption     cmp.Option
		ExpectedOuput []action.Action
		ExpectedError error
	}{
		{
			Name: "Mount Block Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.MountDeviceAction{},
				service.LinuxDeviceService{},
			),
			ExpectedOuput: []action.Action{
				action.NewMountDeviceAction("/dev/xvdf", "/mnt/foo", model.Ext4, "nouuid", nil).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Block Device Already Mounted to Mount Point",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					MountPoint: "/mnt/foo",
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Unmount Block Device From Existing Location and Mount Block Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					MountPoint: "/mnt/bar",
					FileSystem: model.Ext4,
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.UnmountDeviceAction{},
				action.MountDeviceAction{},
				service.LinuxDeviceService{},
			),
			ExpectedOuput: []action.Action{
				action.NewUnmountDeviceAction("/dev/xvdf", "/mnt/bar", nil).SetMode(config.DefaultMode),
				action.NewMountDeviceAction("/dev/xvdf", "/mnt/foo", model.Ext4, "nouuid", nil).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Remount Block Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							Remount:      true,
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					MountPoint: "/mnt/foo",
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.MountDeviceAction{},
				service.LinuxDeviceService{},
			),
			ExpectedOuput: []action.Action{
				action.NewMountDeviceAction("/dev/xvdf", "/mnt/foo", model.Ext4, "nouuid,remount", nil).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Mount Block Device to Symbolic Link",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/bar",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Files: map[string]*model.File{
				"/mnt/bar": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.MountDeviceAction{},
				service.LinuxDeviceService{},
			),
			ExpectedOuput: []action.Action{
				action.NewMountDeviceAction("/dev/xvdf", "/mnt/bar", model.Ext4, "nouuid", nil).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Block Device Already Mounted to Symbolic Linked Directory",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/bar",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					MountPoint: "/mnt/foo",
				},
			},
			Files: map[string]*model.File{
				"/mnt/bar": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Invalid + Block Device Does Not Have File System",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Unformatted,
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: nil,
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Can not mount a device with no file system"),
		},
		{
			Name: "Invalid + Mount Point Does Not Exist",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Files: map[string]*model.File{
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: nil,
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: /mnt/foo must exist as a directory before it can be mounted"),
		},
		{
			Name: "Invalid + Mount Point Is Already Mounted",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1500,
					InodeNo:  2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2500,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: nil,
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: /mnt/foo is already mounted by another device"),
		},
		{
			Name: "Skip + Mount Point Not Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Files:         map[string]*model.File{},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			mdl := NewMountDeviceLayer(ldb, lfb)
			actions, err := mdl.Modify(subtest.Config)
			utils.CheckError("mdl.Modify()", t, subtest.ExpectedError, err)

			utils.CheckOutput("mdl.Modify()", t, subtest.ExpectedOuput, actions, subtest.CmpOption)
		})
	}
}

func TestMountDeviceLayerValidate(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Devices       map[string]*model.BlockDevice
		Files         map[string]*model.File
		ExpectedError error
	}{
		{
			Name: "Block Device Mounted to Mount Point",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					MountPoint: "/mnt/foo",
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Block Device Mounted to Symbolic Link",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/bar",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					MountPoint: "/mnt/foo",
				},
			},
			Files: map[string]*model.File{
				"/mnt/bar": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid + Mount Point Does Not Exist",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					MountPoint: "/mnt/foo",
				},
			},
			Files:         map[string]*model.File{},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed ownership validation checks. /mnt/foo is either not a directory or does not exist"),
		},
		{
			Name: "Invalid + Block Device Not Mounted to Requested Location",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Ext4,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							MountOptions: "nouuid",
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					MountPoint: "/mnt/bar",
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed mountpoint validation checks. Device not mounted to /mnt/foo"),
		},
		{
			Name: "Skip + Mount Point Not Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Files:         map[string]*model.File{},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			mdl := NewMountDeviceLayer(ldb, lfb)
			err := mdl.Validate(subtest.Config)
			utils.CheckError("mdl.Validate()", t, subtest.ExpectedError, err)
		})
	}
}
