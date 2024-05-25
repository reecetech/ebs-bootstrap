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
			Name: "Resize Not Required: File System Size Within FileSystemResizeThreshold",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							Resize: true,
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
					FileSystemSize:  999990,
					BlockDeviceSize: 1000000,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Resize Block Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							Resize: true,
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
					FileSystemSize:  999989,
					BlockDeviceSize: 1000000,
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
		// These devices must be mounted prior to a resize operation.
		// The target of the resize opeartion would be the mountpoint, rather
		// than the device itself. XFS is a file system produces this behaviour
		{
			Name: "Resize Block Device: Must be Mounted",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs:         model.Xfs,
						MountPoint: "/mnt/foo",
						Options: config.Options{
							Resize: true,
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
					FileSystemSize:  999989,
					BlockDeviceSize: 1000000,
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
			Name: "Resize Block Device: Must be Mounted But No Mount Point Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Xfs,
						Options: config.Options{
							Resize: true,
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
					FileSystemSize:  999989,
					BlockDeviceSize: 1000000,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: nil,
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: To resize the xfs file system, device must be mounted"),
		},
		{
			Name: "Skip + Resize Disabled",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						Fs: model.Ext4,
						Options: config.Options{
							Resize: false,
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
					FileSystemSize:  999989,
					BlockDeviceSize: 1000000,
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
							Resize: true,
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
					FileSystemSize:  999990,
					BlockDeviceSize: 1000000,
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
							Resize: false,
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
					FileSystemSize:  999989,
					BlockDeviceSize: 1000000,
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
							Resize: true,
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
					FileSystemSize:  999989,
					BlockDeviceSize: 1000000,
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed to resize file system. File System=999989 Block Device=1000000 (bytes)"),
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

func TestResizeDeviceLayerShouldProcess(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		ExpectedOutput bool
	}{
		{
			Name: "At Least Once Device Has Resize Enabled",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						Options: config.Options{
							Resize: true,
						},
					},
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: true,
		},
		{
			Name: "No Device Has Resize Enabled",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: false,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			rl := NewResizeDeviceLayer(nil, nil)
			output := rl.ShouldProcess(subtest.Config)
			utils.CheckOutput("rl.ShouldProcess()", t, subtest.ExpectedOutput, output)
		})
	}
}
