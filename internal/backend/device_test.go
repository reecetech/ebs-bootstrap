package backend

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestGetBlockDevice(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		Device         string
		BlockDevices   map[string]*model.BlockDevice
		ExpectedOutput *model.BlockDevice
		ExpectedError  error
	}{
		{
			Name:   "Valid Block Device",
			Device: "/dev/xvdf",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			ExpectedOutput: &model.BlockDevice{
				Name:       "/dev/xvdf",
				FileSystem: model.Xfs,
			},
			ExpectedError: nil,
		},
		{
			Name:   "Non-existent Block Device",
			Device: "/dev/sdb",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/sdb: Could not find block device"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxDeviceBackend(subtest.BlockDevices)
			bd, err := ldb.GetBlockDevice(subtest.Device)
			utils.CheckError("ldb.From()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ldb.From()", t, subtest.ExpectedOutput, bd)
		})
	}
}

func TestLabel(t *testing.T) {
	subtests := []struct {
		Name           string
		BlockDevices   map[string]*model.BlockDevice
		Device         string
		Label          string
		CmpOption      cmp.Option
		ExpectedOutput []action.Action
		ExpectedError  error
	}{
		{
			Name: "Labelling Valid Block Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Device: "/dev/xvdf",
			Label:  "label",
			CmpOption: cmp.AllowUnexported(
				action.LabelDeviceAction{},
				service.Ext4Service{},
			),
			ExpectedOutput: []action.Action{
				action.NewLabelDeviceAction("/dev/xvdf", "label", service.NewExt4Service(nil)),
			},
			ExpectedError: nil,
		},
		{
			Name: "Labelling Valid Block Device + Requires Unmount",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
					MountPoint: "/mnt/app",
				},
			},
			Device: "/dev/xvdf",
			Label:  "label",
			CmpOption: cmp.AllowUnexported(
				action.LabelDeviceAction{},
				action.UnmountDeviceAction{},
				service.XfsService{},
			),
			ExpectedOutput: []action.Action{
				action.NewUnmountDeviceAction("/dev/xvdf", "/mnt/app", nil),
				action.NewLabelDeviceAction("/dev/xvdf", "label", service.NewXfsService(nil)),
			},
			ExpectedError: nil,
		},
		{
			Name: "Fail To Label + Exceeding Maximum Label Length",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Device: "/dev/xvdf",
			Label:  "exceptionally-long-label",
			CmpOption: cmp.AllowUnexported(
				action.LabelDeviceAction{},
				service.Ext4Service{},
			),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: Label 'exceptionally-long-label' exceeds the maximum 16 character length for the ext4 file system"),
		},
		{
			Name: "Fail To Label + Unformatted Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Unformatted,
				},
			},
			Device: "/dev/xvdf",
			Label:  "label",
			CmpOption: cmp.AllowUnexported(
				action.LabelDeviceAction{},
			),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: An unformatted file system can not be queried/modified"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxDeviceBackend(subtest.BlockDevices)
			bd, err := ldb.GetBlockDevice(subtest.Device)
			utils.ExpectErr("ldb.GetBlockDevice()", t, false, err)

			actions, err := ldb.Label(bd, subtest.Label)
			utils.CheckError("ldb.Label()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ldb.Label()", t, subtest.ExpectedOutput, actions, subtest.CmpOption)
		})
	}
}

func TestResize(t *testing.T) {
	subtests := []struct {
		Name           string
		BlockDevices   map[string]*model.BlockDevice
		Device         string
		CmpOption      cmp.Option
		ExpectedOutput action.Action
		ExpectedError  error
	}{
		{
			Name: "Valid Block Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			Device: "/dev/xvdf",
			CmpOption: cmp.AllowUnexported(
				service.Ext4Service{},
				action.ResizeDeviceAction{},
			),
			ExpectedOutput: action.NewResizeDeviceAction("/dev/xvdf", "/dev/xvdf", service.NewExt4Service(nil)),
			ExpectedError:  nil,
		},
		{
			Name: "Valid Block Device + Requires Mount",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
					MountPoint: "/mnt/app",
				},
			},
			Device: "/dev/xvdf",
			CmpOption: cmp.AllowUnexported(
				service.XfsService{},
				action.ResizeDeviceAction{},
			),
			ExpectedOutput: action.NewResizeDeviceAction("/dev/xvdf", "/mnt/app", service.NewXfsService(nil)),
			ExpectedError:  nil,
		},
		{
			Name: "Fail to Resize + Requires Mount, but not Mounted",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			Device:         "/dev/xvdf",
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: To resize the xfs file system, device must be mounted"),
		},
		{
			Name: "Fail To Resize + Unformatted Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Unformatted,
				},
			},
			Device:         "/dev/xvdf",
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: An unformatted file system can not be queried/modified"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxDeviceBackend(subtest.BlockDevices)
			bd, err := ldb.GetBlockDevice(subtest.Device)
			utils.ExpectErr("ldb.GetBlockDevice()", t, false, err)

			action, err := ldb.Resize(bd)
			utils.CheckError("ldb.Resize()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ldb.Resize()", t, subtest.ExpectedOutput, action, subtest.CmpOption)
		})
	}
}

func TestFormat(t *testing.T) {
	subtests := []struct {
		Name         string
		BlockDevices map[string]*model.BlockDevice
		Device       string
		model.FileSystem
		CmpOption      cmp.Option
		ExpectedOutput action.Action
		ExpectedError  error
	}{
		{
			Name: "Valid Block Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Unformatted,
				},
			},
			Device:     "/dev/xvdf",
			FileSystem: model.Ext4,
			CmpOption: cmp.AllowUnexported(
				service.Ext4Service{},
				action.FormatDeviceAction{},
			),
			ExpectedOutput: action.NewFormatDeviceAction("/dev/xvdf", service.NewExt4Service(nil)),
			ExpectedError:  nil,
		},
		{
			Name: "Invalid File System + Attempting to Erase File System",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			Device:         "/dev/xvdf",
			FileSystem:     model.Unformatted,
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: An unformatted file system can not be queried/modified"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxDeviceBackend(subtest.BlockDevices)
			bd, err := ldb.GetBlockDevice(subtest.Device)
			utils.ExpectErr("ldb.GetBlockDevice()", t, false, err)

			action, err := ldb.Format(bd, subtest.FileSystem)
			utils.CheckError("ldb.Format()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ldb.Format()", t, subtest.ExpectedOutput, action, subtest.CmpOption)
		})
	}
}

func TestMount(t *testing.T) {
	subtests := []struct {
		Name         string
		BlockDevices map[string]*model.BlockDevice
		Device       string
		Target       string
		model.MountOptions
		CmpOption      cmp.Option
		ExpectedOutput action.Action
	}{
		{
			Name: "Valid Block Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			Device:         "/dev/xvdf",
			Target:         "/mnt/app",
			MountOptions:   model.MountOptions("defaults"),
			CmpOption:      cmp.AllowUnexported(action.MountDeviceAction{}),
			ExpectedOutput: action.NewMountDeviceAction("/dev/xvdf", "/mnt/app", model.Xfs, model.MountOptions("defaults"), nil),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxDeviceBackend(subtest.BlockDevices)
			bd, err := ldb.GetBlockDevice(subtest.Device)
			utils.ExpectErr("ldb.GetBlockDevice()", t, false, err)

			action := ldb.Mount(bd, subtest.Target, subtest.MountOptions)
			utils.CheckOutput("ldb.Format()", t, subtest.ExpectedOutput, action, subtest.CmpOption)
		})
	}
}

func TestRemount(t *testing.T) {
	subtests := []struct {
		Name         string
		BlockDevices map[string]*model.BlockDevice
		Device       string
		Target       string
		model.MountOptions
		CmpOption      cmp.Option
		ExpectedOutput action.Action
	}{
		{
			Name: "Valid Block Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			Device:         "/dev/xvdf",
			Target:         "/mnt/app",
			MountOptions:   model.MountOptions("defaults"),
			CmpOption:      cmp.AllowUnexported(action.MountDeviceAction{}),
			ExpectedOutput: action.NewMountDeviceAction("/dev/xvdf", "/mnt/app", model.Xfs, model.MountOptions("defaults,remount"), nil),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxDeviceBackend(subtest.BlockDevices)
			bd, err := ldb.GetBlockDevice(subtest.Device)
			utils.ExpectErr("ldb.GetBlockDevice()", t, false, err)

			action := ldb.Remount(bd, subtest.Target, subtest.MountOptions)
			utils.CheckOutput("ldb.Format()", t, subtest.ExpectedOutput, action, subtest.CmpOption)
		})
	}
}

func TestUnmount(t *testing.T) {
	subtests := []struct {
		Name           string
		BlockDevices   map[string]*model.BlockDevice
		Device         string
		CmpOption      cmp.Option
		ExpectedOutput action.Action
	}{
		{
			Name: "Valid Block Device",
			BlockDevices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
					MountPoint: "/mnt/app",
				},
			},
			Device:         "/dev/xvdf",
			CmpOption:      cmp.AllowUnexported(action.UnmountDeviceAction{}),
			ExpectedOutput: action.NewUnmountDeviceAction("/dev/xvdf", "/mnt/app", nil),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxDeviceBackend(subtest.BlockDevices)
			bd, err := ldb.GetBlockDevice(subtest.Device)
			utils.ExpectErr("ldb.GetBlockDevice()", t, false, err)

			action := ldb.Umount(bd)
			utils.CheckOutput("ldb.Format()", t, subtest.ExpectedOutput, action, subtest.CmpOption)
		})
	}
}

func TestLinuxDeviceBackendFrom(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		GetBlockDevice func(name string) (*model.BlockDevice, error)
		ExpectedOutput map[string]*model.BlockDevice
		ExpectedError  error
	}{
		{
			Name: "Valid Block Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			GetBlockDevice: func(name string) (*model.BlockDevice, error) {
				return &model.BlockDevice{
					Name:       name,
					FileSystem: model.Unformatted,
				}, nil
			},
			ExpectedOutput: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Unformatted,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid Block Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			GetBlockDevice: func(name string) (*model.BlockDevice, error) {
				return nil, fmt.Errorf("🔴 %s: 'jfs' is not a supported file system", name)
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: 'jfs' is not a supported file system"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mds := service.NewMockDeviceService()
			if subtest.GetBlockDevice != nil {
				mds.StubGetBlockDevice = subtest.GetBlockDevice
			}

			ldb := NewLinuxDeviceBackend(mds, nil)
			err := ldb.From(subtest.Config)
			utils.CheckError("ldb.From()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ldb.From()", t, subtest.ExpectedOutput, ldb.blockDevices)
		})
	}
}
