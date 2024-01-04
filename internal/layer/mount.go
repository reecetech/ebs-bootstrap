package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
)

type MountDeviceLayer struct {
	deviceBackend backend.DeviceBackend
	fileBackend   backend.FileBackend
}

func NewMountDeviceLayer(db backend.DeviceBackend, fb backend.FileBackend) *MountDeviceLayer {
	return &MountDeviceLayer{
		deviceBackend: db,
		fileBackend:   fb,
	}
}

func (fdl *MountDeviceLayer) From(c *config.Config) error {
	err := fdl.deviceBackend.From(c)
	if err != nil {
		return err
	}
	return fdl.fileBackend.From(c)
}

func (fdl *MountDeviceLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}

		bd, err := fdl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return nil, err
		}
		if bd.FileSystem == model.Unformatted {
			return nil, fmt.Errorf("🔴 %s: Can not mount a device with no file system", bd.Name)
		}

		d, err := fdl.fileBackend.GetDirectory(cd.MountPoint)
		if err != nil {
			return nil, fmt.Errorf("🔴 %s: %s must exist as a directory before it can be mounted", name, cd.MountPoint)
		}

		mode := c.GetMode(name)
		mo := c.GetMountOptions(name)
		if bd.MountPoint == d.Path {
			if c.GetRemount(name) {
				a := fdl.deviceBackend.Remount(bd, cd.MountPoint, mo).SetMode(mode)
				actions = append(actions, a)
			}
		} else {
			if fdl.fileBackend.IsMount(cd.MountPoint) {
				return nil, fmt.Errorf("🔴 %s: %s is already mounted by another device", name, cd.MountPoint)
			}
			// If mount point already exists, then lets unmount it first
			if len(bd.MountPoint) > 0 {
				a := fdl.deviceBackend.Umount(bd).SetMode(mode)
				actions = append(actions, a)
			}
			a := fdl.deviceBackend.Mount(bd, cd.MountPoint, mo).SetMode(mode)
			actions = append(actions, a)
		}
	}
	return actions, nil
}

func (fdl *MountDeviceLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}
		bd, err := fdl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return err
		}
		d, err := fdl.fileBackend.GetDirectory(cd.MountPoint)
		if err != nil {
			return fmt.Errorf("🔴 %s: Failed ownership validation checks. %s is either not a directory or does not exist", name, cd.MountPoint)
		}
		if bd.MountPoint != d.Path {
			return fmt.Errorf("🔴 %s: Failed mountpoint validation checks. Device not mounted to %s", name, cd.MountPoint)
		}
	}
	return nil
}

func (fdl *MountDeviceLayer) Warning() string {
	return "Devices mounted to a location, not specified in the configuration, will be unmounted"
}
