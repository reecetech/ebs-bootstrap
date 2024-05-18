package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
)

type CreatePhysicalVolumeLayer struct {
	deviceBackend backend.DeviceBackend
	lvmBackend    backend.LvmBackend
}

func NewCreatePhysicalVolumeLayer(db backend.DeviceBackend, lb backend.LvmBackend) *CreatePhysicalVolumeLayer {
	return &CreatePhysicalVolumeLayer{
		deviceBackend: db,
		lvmBackend:    lb,
	}
}

func (cpvl *CreatePhysicalVolumeLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		bd, err := cpvl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return nil, err
		}
		if bd.FileSystem == model.Lvm {
			continue
		}
		if bd.FileSystem != model.Unformatted {
			return nil, fmt.Errorf("🔴 %s: Can not create a physical volume on a device with an existing %s file system", bd.Name, bd.FileSystem.String())
		}
		mode := c.GetMode(name)
		a := cpvl.lvmBackend.CreatePhysicalVolume(bd.Name)
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (cpvl *CreatePhysicalVolumeLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		bd, err := cpvl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return err
		}
		if bd.FileSystem != model.Lvm {
			return fmt.Errorf("🔴 %s: Failed physical volume validation checks. Expected=%s, Actual=%s", bd.Name, model.Lvm, bd.FileSystem)
		}
	}
	return nil
}

func (cpvl *CreatePhysicalVolumeLayer) Warning() string {
	return DisabledWarning
}

func (cpvl *CreatePhysicalVolumeLayer) From(c *config.Config) error {
	// Lvmackend does not have to be initialised as we are only creating a Physical Volume
	return cpvl.deviceBackend.From(c)
}
