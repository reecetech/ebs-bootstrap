package backend

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

const (
	// The % threshold at which to resize a file system
	// -------------------------------------------------------
	// If the (file system size / device size) * 100 falls
	// under this threshold then we perform a resize operation
	// -------------------------------------------------------
	// Why is the threshold not set to 100%?
	//	- A completely extended file system may be size that is
	//	  slightly less than that of the underlying block device
	//	- This is likely due to reserved sections that store
	//	  file system metadata
	//	- Therefore we set the threshold to 99.999% to avoid
	//	  unnecessary resize operations
	// Why is the threshold set to 99.999%?
	//	- The largest EBS volume you can provision is a 64 TiB
	//	  io2 Block Express volume.
	//	- EBS volumes can only be specified in increments of 1 GiB
	//	- 64 TiB = 65536 GiB | 99.999 % * 65536 = 65535.34 GiB
	//	- Therefore, a resize threshold of 99.999% ensures a resize
	//	  operation from 65535 GiB to 65536 GiB, since 65535 < 65535.34
	FileSystemResizeThreshold = float64(99.999)
)

type DeviceMetricsBackend interface {
	GetBlockDeviceMetrics(name string) (*model.BlockDeviceMetrics, error)
	ShouldResize(bdm *model.BlockDeviceMetrics) bool
	From(config *config.Config) error
}

type LinuxDeviceMetricsBackend struct {
	blockDeviceMetrics       map[string]*model.BlockDeviceMetrics
	deviceService            service.DeviceService
	fileSystemServiceFactory service.FileSystemServiceFactory
}

func NewLinuxDeviceMetricsBackend(ds service.DeviceService, fssf service.FileSystemServiceFactory) *LinuxDeviceMetricsBackend {
	return &LinuxDeviceMetricsBackend{
		blockDeviceMetrics:       map[string]*model.BlockDeviceMetrics{},
		deviceService:            ds,
		fileSystemServiceFactory: fssf,
	}
}

func NewMockLinuxDeviceMetricsBackend(blockDeviceMetrics map[string]*model.BlockDeviceMetrics) *LinuxDeviceMetricsBackend {
	return &LinuxDeviceMetricsBackend{
		blockDeviceMetrics:       blockDeviceMetrics,
		deviceService:            nil,
		fileSystemServiceFactory: service.NewLinuxFileSystemServiceFactory(nil),
	}
}

func (dmb *LinuxDeviceMetricsBackend) GetBlockDeviceMetrics(name string) (*model.BlockDeviceMetrics, error) {
	metrics, exists := dmb.blockDeviceMetrics[name]
	if !exists {
		return nil, fmt.Errorf("🔴 %s: Could not find block device metrics", name)
	}
	return metrics, nil
}

func (dmb *LinuxDeviceMetricsBackend) ShouldResize(bdm *model.BlockDeviceMetrics) bool {
	return (float64(bdm.FileSystemSize) / float64(bdm.BlockDeviceSize) * 100) < FileSystemResizeThreshold
}

func (dmb *LinuxDeviceMetricsBackend) From(config *config.Config) error {
	dmb.blockDeviceMetrics = nil
	blockDeviceMetrics := map[string]*model.BlockDeviceMetrics{}

	for name := range config.Devices {
		bd, err := dmb.deviceService.GetBlockDevice(name)
		if err != nil {
			return err
		}
		fs, err := dmb.fileSystemServiceFactory.Select(bd.FileSystem)
		if err != nil {
			return fmt.Errorf("🔴 %s: %s", bd.Name, err)
		}
		// Block Device Size
		bss, err := dmb.deviceService.GetSize(bd.Name)
		if err != nil {
			return err
		}
		// File System Size
		fss, err := fs.GetSize(bd.Name)
		if err != nil {
			return err
		}
		blockDeviceMetrics[bd.Name] = &model.BlockDeviceMetrics{
			BlockDeviceSize: bss,
			FileSystemSize:  fss,
		}
	}
	dmb.blockDeviceMetrics = blockDeviceMetrics
	return nil
}
