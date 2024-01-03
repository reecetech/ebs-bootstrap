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

func TestLabelDeviceLayerModify(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Devices       map[string]*model.BlockDevice
		CmpOption     cmp.Option
		ExpectedOuput []action.Action
		ExpectedError error
	}{
		{
			Name: "Label Device With File System That Avoids Unmounting",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:    model.Ext4,
						Label: "label",
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.LabelDeviceAction{},
				service.Ext4Service{},
			),
			ExpectedOuput: []action.Action{
				action.NewLabelDeviceAction("/dev/xvdf", "label", service.NewExt4Service(nil)).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Label Device With File System That Requires Unmounting",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Xfs,
						Label:      "label",
						MountPoint: "/mnt/foo",
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
					MountPoint: "/mnt/foo",
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.UnmountDeviceAction{},
				action.LabelDeviceAction{},
				service.XfsService{},
			),
			ExpectedOuput: []action.Action{
				action.NewUnmountDeviceAction("/dev/xvdf", "/mnt/foo", nil).SetMode(config.DefaultMode),
				action.NewLabelDeviceAction("/dev/xvdf", "label", service.NewXfsService(nil)).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Label Matches Requested Label",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:    model.Ext4,
						Label: "label",
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					Label:      "label",
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Skip Labelling",
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
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Attempting to Label Block Device With No File System",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:    model.Ext4,
						Label: "label",
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name: "/dev/xvdf",
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: nil,
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: An unformatted file system can not be queried/modified"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			layer := NewLabelDeviceLayer(ldb)
			actions, err := layer.Modify(subtest.Config)
			utils.CheckError("LabelDeviceLayer.Modify()", t, subtest.ExpectedError, err)
			utils.CheckOutput("LabelDeviceLayer.Modify()", t, subtest.ExpectedOuput, actions, subtest.CmpOption)
		})
	}
}

func TestLabelDeviceLayerValidate(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Devices       map[string]*model.BlockDevice
		ExpectedError error
	}{
		{
			Name: "Label Matches Requested Label",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:    model.Ext4,
						Label: "label",
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					Label:      "label",
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Skipping Validation",
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
			ExpectedError: nil,
		},
		{
			Name: "Label Does Not Match Requested Label",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:    model.Ext4,
						Label: "label",
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Ext4,
					Label:      "not-label",
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed label validation checks. Expected=label, Actual=not-label"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			ldl := NewLabelDeviceLayer(ldb)
			err := ldl.Validate(subtest.Config)
			utils.CheckError("ldl.Validate()", t, subtest.ExpectedError, err)
		})
	}
}
