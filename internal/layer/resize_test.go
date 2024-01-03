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

func TestResizeDeviceLayerModify(t *testing.T) {
	subtests := []struct {
		Name               string
		Config             *config.Config
		Devices            map[string]*model.BlockDevice
		BlockDeviceMetrics map[string]*model.BlockDeviceMetrics
		CmpOption          cmp.Option
		ExpectedOuput      []action.Action
		ExpectedError      error
	}{
		{
			Name: "Resize Not Required: File System Size Within Threshold",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 90,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  95,
					BlockDeviceSize: 100,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Resize Block Device + Resize Threshold=90% + Target=Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 90,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.ResizeDeviceAction{},
				service.Ext4Service{},
			),
			ExpectedOuput: []action.Action{
				action.NewResizeDeviceAction("/dev/xvdf", "/dev/xvdf", service.NewExt4Service(nil)).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Resize Block Device + Resize Threshold=90% + Target=Mount Point",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Xfs,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 90,
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					MountPoint: "/mnt/foo",
					FileSystem: model.Xfs,
				},
			},
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.ResizeDeviceAction{},
				service.XfsService{},
			),
			ExpectedOuput: []action.Action{
				action.NewResizeDeviceAction("/dev/xvdf", "/mnt/foo", service.NewXfsService(nil)).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Resize Block Device (With Mount Point) + But No Mount Point Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Xfs,
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 90,
						},
					},
				},
			},
			Devices: map[string]*model.BlockDevice{
				"/dev/xvdf": {
					Name:       "/dev/xvdf",
					FileSystem: model.Xfs,
				},
			},
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: nil,
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: To resize the xfs file system, device must be mounted"),
		},
		{
			Name: "Resize Block Device + Resize Threshold=0% (Always Resize) + Target=Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 0,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  100,
					BlockDeviceSize: 100,
				},
			},
			CmpOption: cmp.AllowUnexported(
				action.ResizeDeviceAction{},
				service.Ext4Service{},
			),
			ExpectedOuput: []action.Action{
				action.NewResizeDeviceAction("/dev/xvdf", "/dev/xvdf", service.NewExt4Service(nil)).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Skip + Resize Disabled",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs: false,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			lbdb := backend.NewMockLinuxDeviceMetricsBackend(subtest.BlockDeviceMetrics)
			rl := NewResizeDeviceLayer(ldb, lbdb)
			actions, err := rl.Modify(subtest.Config)
			utils.CheckError("rl.Modify()", t, subtest.ExpectedError, err)
			utils.CheckOutput("rl.Modify()", t, subtest.ExpectedOuput, actions, subtest.CmpOption)
		})
	}
}

func TestResizeDeviceLayerValidate(t *testing.T) {
	subtests := []struct {
		Name               string
		Config             *config.Config
		Devices            map[string]*model.BlockDevice
		BlockDeviceMetrics map[string]*model.BlockDeviceMetrics
		ExpectedError      error
	}{
		{
			Name: "File System Size Within Resize Threshold",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 90,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  95,
					BlockDeviceSize: 100,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Skip + Resize Threshold=0%",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 0,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Skip + Resize Disabled",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs: false,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid + Resize Still Expected",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							ResizeFs:        true,
							ResizeThreshold: 90,
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
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed to resize file system. File System=80 Block Device=100 (bytes)"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := backend.NewMockLinuxDeviceBackend(subtest.Devices)
			lbdb := backend.NewMockLinuxDeviceMetricsBackend(subtest.BlockDeviceMetrics)
			rl := NewResizeDeviceLayer(ldb, lbdb)
			err := rl.Validate(subtest.Config)
			utils.CheckError("rl.Modify()", t, subtest.ExpectedError, err)
		})
	}
}
