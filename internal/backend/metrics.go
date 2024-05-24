package backend

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type DeviceMetricsBackend interface {
	GetBlockDeviceMetrics(name string) (*model.BlockDeviceMetrics, error)
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
