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

func TestFormatDeviceLayerModify(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		Devices        map[string]*model.BlockDevice
		CmpOption      cmp.Option
		ExpectedOutput []action.Action
		ExpectedError  error
	}{
		{
			Name: "Formatting Unformatted Block Device to XFS",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Xfs,
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Unformatted,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.FormatDeviceAction{},
				service.XfsService{},
			),
			ExpectedOutput: []action.Action{
				action.NewFormatDeviceAction("/dev/xvdf", service.NewXfsService(nil)).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "File System Matches Requested File System",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Xfs,
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: []action.Action{},
			ExpectedError:  nil,
		},
		{
			Name: "Attempting to Erase the XFS File System of Block Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Unformatted,
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: Can not erase the file system of a device"),
		},
		{
			Name: "Attempting to Change the Existing File System of a Block Device",
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
					FileSystem: model.Xfs,
				},
			},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: Can not format a device with an existing xfs file system"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			ld := NewFormatDeviceLayer(ldb)
			actions, err := ld.Modify(subtest.Config)
			utils.CheckError("ld.Modify()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ld.Modify()", t, subtest.ExpectedOutput, actions, subtest.CmpOption)
		})
	}
}

func TestFormatDeviceLayerValidate(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		Devices        map[string]*model.BlockDevice
		ExpectedOutput error
	}{
		{
			Name: "File System Matches Requested File System",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Xfs,
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			ExpectedOutput: nil,
		},
		{
			Name: "File System Does Not Match Requested File System",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Xfs,
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Unformatted,
				},
			},
			ExpectedOutput: fmt.Errorf("🔴 /dev/xvdf: Failed file system validation checks. Expected=xfs, Actual=unformatted"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			ld := NewFormatDeviceLayer(ldb)
			err := ld.Validate(subtest.Config)
			utils.CheckError("ld.Validate()", t, subtest.ExpectedOutput, err)
		})
	}
}
