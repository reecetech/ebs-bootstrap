package backend

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestGetBlockDeviceMetrics(t *testing.T) {
	subtests := []struct {
		Name               string
		Device             string
		BlockDeviceMetrics map[string]*model.BlockDeviceMetrics
		ExpectedOutput     *model.BlockDeviceMetrics
		ExpectedError      error
	}{
		{
			Name:   "Valid Device",
			Device: "/dev/xvdf",
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			ExpectedOutput: &model.BlockDeviceMetrics{
				FileSystemSize:  80,
				BlockDeviceSize: 100,
			},
			ExpectedError: nil,
		},
		{
			Name:               "Invalid Device",
			Device:             "/dev/xvdf",
			BlockDeviceMetrics: map[string]*model.BlockDeviceMetrics{},
			ExpectedOutput:     nil,
			ExpectedError:      fmt.Errorf("🔴 /dev/xvdf: Could not find block device metrics"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			dmb := NewMockLinuxDeviceMetricsBackend(subtest.BlockDeviceMetrics)
			metrics, err := dmb.GetBlockDeviceMetrics(subtest.Device)
			utils.CheckError("dmb.GetBlockDeviceMetrics()", t, subtest.ExpectedError, err)
			utils.CheckOutput("dmb.GetBlockDeviceMetrics()", t, subtest.ExpectedOutput, metrics)
		})
	}
}

func TestLinuxDeviceMetricsBackendFrom(t *testing.T) {
	fssf := service.NewLinuxFileSystemServiceFactory(nil)

	subtests := []struct {
		Name              string
		Config            *config.Config
		GetBlockDevice    func(name string) (*model.BlockDevice, error)
		GetDeviceSize     func(name string) (uint64, error)
		GetFileSystemSize func(name string) (uint64, error)
		ExpectedOutput    map[string]*model.BlockDeviceMetrics
		ExpectedError     error
	}{
		{
			Name: "Valid Device",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			GetBlockDevice: func(name string) (*model.BlockDevice, error) {
				return &model.BlockDevice{
					Name:       name,
					FileSystem: model.Ext4,
				}, nil
			},
			GetDeviceSize: func(name string) (uint64, error) {
				return 100, nil
			},
			GetFileSystemSize: func(name string) (uint64, error) {
				return 80, nil
			},
			ExpectedOutput: map[string]*model.BlockDeviceMetrics{
				"/dev/xvdf": {
					FileSystemSize:  80,
					BlockDeviceSize: 100,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Unsupported File System",
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
		{
			Name: "Unformatted File System",
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
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /dev/xvdf: An unformatted file system can not be queried/modified"),
		},
		{
			Name: "Failure to Get Device Size",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			GetBlockDevice: func(name string) (*model.BlockDevice, error) {
				return &model.BlockDevice{
					Name:       name,
					FileSystem: model.Ext4,
				}, nil
			},
			GetDeviceSize: func(name string) (uint64, error) {
				return 0, fmt.Errorf("🔴 blockdev is either not installed or accessible from $PATH")
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 blockdev is either not installed or accessible from $PATH"),
		},
		{
			Name: "Failure to Get File System Size",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			GetBlockDevice: func(name string) (*model.BlockDevice, error) {
				return &model.BlockDevice{
					Name:       name,
					FileSystem: model.Ext4,
				}, nil
			},
			GetDeviceSize: func(name string) (uint64, error) {
				return 100, nil
			},
			GetFileSystemSize: func(name string) (uint64, error) {
				return 0, fmt.Errorf("🔴 tune2fs is either not installed or accessible from $PATH")
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 tune2fs is either not installed or accessible from $PATH"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ds := service.NewMockDeviceService()
			if subtest.GetBlockDevice != nil {
				ds.StubGetBlockDevice = subtest.GetBlockDevice
			}
			if subtest.GetDeviceSize != nil {
				ds.StubGetSize = subtest.GetDeviceSize
			}
			fss := service.NewMockFileSystemService()
			if subtest.GetFileSystemSize != nil {
				fss.StubGetSize = subtest.GetFileSystemSize
			}

			fssf := service.NewMockFileSystemServiceFactory(fssf, fss)
			dmb := NewLinuxDeviceMetricsBackend(ds, fssf)

			err := dmb.From(subtest.Config)
			utils.CheckError("dmb.From()", t, subtest.ExpectedError, err)
			utils.CheckOutput("dmb.From()", t, subtest.ExpectedOutput, dmb.blockDeviceMetrics)
		})
	}
}
