package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type CreateLogicalVolumeLayer struct {
	lvmBackend backend.LvmBackend
}

func NewCreateLogicalVolumeLayer(lb backend.LvmBackend) *CreateLogicalVolumeLayer {
	return &CreateLogicalVolumeLayer{
		lvmBackend: lb,
	}
}

func (cvgl *CreateLogicalVolumeLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}

		lvs := cvgl.lvmBackend.SearchLogicalVolumes(cd.Lvm)
		if len(lvs) == 1 {
			if lvs[0].Name == cd.Lvm {
				continue
			}
			return nil, fmt.Errorf("🔴 %s: Volume group %s already has logical volume %s associated", name, cd.Lvm, lvs[0].Name)
		}
		if len(lvs) > 1 {
			return nil, fmt.Errorf("🔴 %s: Cannot manage volume group %s with more than one logical volume associated", name, cd.Lvm)
		}

		mode := c.GetMode(name)
		a := cvgl.lvmBackend.CreateLogicalVolume(cd.Lvm, cd.Lvm, c.GetLvmConsumption(name))
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (cvgl *CreateLogicalVolumeLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		lvs := cvgl.lvmBackend.SearchLogicalVolumes(cd.Lvm)
		if len(lvs) == 1 {
			if lvs[0].Name == cd.Lvm {
				continue
			}
			return fmt.Errorf("🔴 %s: Failed to validate logical volume. Expected=%s, Actual=%s", name, cd.Lvm, lvs[0].Name)
		}
		if len(lvs) > 1 {
			return fmt.Errorf("🔴 %s: ailed to validate logical volume. #(Logical Volume) Expected=%d, Actual=%d", name, 1, len(lvs))
		}
	}
	return nil
}

func (cvgl *CreateLogicalVolumeLayer) Warning() string {
	return DisabledWarning
}

func (cvgl *CreateLogicalVolumeLayer) From(c *config.Config) error {
	return cvgl.lvmBackend.From(c)
}

func (cvgl *CreateLogicalVolumeLayer) ShouldProcess(c *config.Config) bool {
	for _, cd := range c.Devices {
		if len(cd.Lvm) > 0 {
			return true
		}
	}
	return false
}
