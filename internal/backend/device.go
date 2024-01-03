package backend

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type DeviceBackend interface {
	GetBlockDevice(device string) (*model.BlockDevice, error)
	Label(bd *model.BlockDevice, label string) ([]action.Action, error)
	Resize(bd *model.BlockDevice) (action.Action, error)
	Format(bd *model.BlockDevice, fileSystem model.FileSystem) (action.Action, error)
	Mount(bd *model.BlockDevice, target string, options model.MountOptions) action.Action
	Remount(bd *model.BlockDevice, target string, options model.MountOptions) action.Action
	Umount(bd *model.BlockDevice) action.Action
	From(config *config.Config) error
}

type LinuxDeviceBackend struct {
	blockDevices             map[string]*model.BlockDevice
	deviceService            service.DeviceService
	fileSystemServiceFactory service.FileSystemServiceFactory
}

func NewLinuxDeviceBackend(ds service.DeviceService, fssf service.FileSystemServiceFactory) *LinuxDeviceBackend {
	return &LinuxDeviceBackend{
		blockDevices:             map[string]*model.BlockDevice{},
		deviceService:            ds,
		fileSystemServiceFactory: fssf,
	}
}

func NewMockLinuxDeviceBackend(blockDevices map[string]*model.BlockDevice) *LinuxDeviceBackend {
	return &LinuxDeviceBackend{
		blockDevices:             blockDevices,
		deviceService:            nil,
		fileSystemServiceFactory: service.NewLinuxFileSystemServiceFactory(nil),
	}
}

func (db *LinuxDeviceBackend) GetBlockDevice(device string) (*model.BlockDevice, error) {
	blockDevice, exists := db.blockDevices[device]
	if !exists {
		return nil, fmt.Errorf("🔴 %s: Could not find block device", device)
	}
	return blockDevice, nil
}

// This method is unique in the sense that some file-systems like xfs might require
// that the device be unmounted before any labelling operations can commence. For these
// file systems, we would preprend any label action with an unmount action (if the device is already mounted)
func (db *LinuxDeviceBackend) Label(bd *model.BlockDevice, label string) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	fss, err := db.fileSystemServiceFactory.Select(bd.FileSystem)
	if err != nil {
		return nil, fmt.Errorf("🔴 %s: %s", bd.Name, err)
	}
	if ml := fss.GetMaximumLabelLength(); len(label) > ml {
		return nil, fmt.Errorf("🔴 %s: Label '%s' exceeds the maximum %d character length for the %s file system", bd.Name, label, ml, fss.GetFileSystem().String())
	}
	if fss.DoesLabelRequireUnmount() && len(bd.MountPoint) > 0 {
		a := db.Umount(bd)
		actions = append(actions, a)
	}
	a := action.NewLabelDeviceAction(
		bd.Name,
		label,
		fss,
	)
	return append(actions, a), nil
}

func (db *LinuxDeviceBackend) Resize(bd *model.BlockDevice) (action.Action, error) {
	fss, err := db.fileSystemServiceFactory.Select(bd.FileSystem)
	if err != nil {
		return nil, fmt.Errorf("🔴 %s: %s", bd.Name, err)
	}
	target := bd.Name
	if fss.DoesResizeRequireMount() {
		if len(bd.MountPoint) == 0 {
			return nil, fmt.Errorf("🔴 %s: To resize the %s file system, device must be mounted", bd.Name, fss.GetFileSystem().String())
		}
		target = bd.MountPoint
	}
	return action.NewResizeDeviceAction(
		bd.Name,
		target,
		fss,
	), nil
}

func (db *LinuxDeviceBackend) Format(bd *model.BlockDevice, fileSystem model.FileSystem) (action.Action, error) {
	fss, err := db.fileSystemServiceFactory.Select(fileSystem)
	if err != nil {
		return nil, fmt.Errorf("🔴 %s: %s", bd.Name, err)
	}
	return action.NewFormatDeviceAction(
		bd.Name,
		fss,
	), nil
}

func (db *LinuxDeviceBackend) Mount(bd *model.BlockDevice, target string, options model.MountOptions) action.Action {
	return action.NewMountDeviceAction(bd.Name, target, bd.FileSystem, options, db.deviceService)
}

func (db *LinuxDeviceBackend) Remount(bd *model.BlockDevice, target string, options model.MountOptions) action.Action {
	return db.Mount(bd, target, options.Remount())
}

func (db *LinuxDeviceBackend) Umount(bd *model.BlockDevice) action.Action {
	return action.NewUnmountDeviceAction(bd.Name, bd.MountPoint, db.deviceService)
}

func (db *LinuxDeviceBackend) From(config *config.Config) error {
	// Clear representation of devices
	db.blockDevices = nil

	blockDevices := map[string]*model.BlockDevice{}
	for name := range config.Devices {
		d, err := db.deviceService.GetBlockDevice(name)
		if err != nil {
			return err
		}
		blockDevices[d.Name] = d
	}
	db.blockDevices = blockDevices
	return nil
}
