package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type CreateVolumeGroupLayer struct {
	lvmBackend backend.LvmBackend
}

func NewCreateVolumeGroupLayer(lb backend.LvmBackend) *CreateVolumeGroupLayer {
	return &CreateVolumeGroupLayer{
		lvmBackend: lb,
	}
}

func (cvgl *CreateVolumeGroupLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		vg := cvgl.lvmBackend.SearchVolumeGroup(name)
		if vg != nil && vg.Name != cd.Lvm {
			return nil, fmt.Errorf("🔴 %s: Physical volume %s already has volume group %s associated", name, name, vg.Name)
		}

		vgs := cvgl.lvmBackend.GetVolumeGroups(cd.Lvm)
		if len(vgs) == 1 {
			if vgs[0].PhysicalVolume == name {
				continue
			}
			return nil, fmt.Errorf("🔴 %s: Volume group %s already exists and belongs to physical volume %s", name, cd.Lvm, vgs[0].PhysicalVolume)
		}
		if len(vgs) > 1 {
			return nil, fmt.Errorf("🔴 %s: Cannot manage volume group %s because it is associated with more than one physical volume", name, cd.Lvm)
		}

		mode := c.GetMode(name)
		a := cvgl.lvmBackend.CreateVolumeGroup(cd.Lvm, name)
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (cvgl *CreateVolumeGroupLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		vgs := cvgl.lvmBackend.GetVolumeGroups(cd.Lvm)
		if len(vgs) == 1 {
			if vgs[0].PhysicalVolume == name {
				return nil
			}
			return fmt.Errorf("🔴 %s: Failed to validate volume group. Expected=%s, Actual=%s", name, name, vgs[0].PhysicalVolume)
		}
		if len(vgs) > 1 {
			return fmt.Errorf("🔴 %s: Failed to validate volume group. #(Physical volume) Expected=%d, Actual=%d", name, 1, len(vgs))
		}

	}
	return nil
}

func (cvgl *CreateVolumeGroupLayer) Warning() string {
	return DisabledWarning
}

func (cvgl *CreateVolumeGroupLayer) From(c *config.Config) error {
	return cvgl.lvmBackend.From(c)
}

func (cvgl *CreateVolumeGroupLayer) ShouldProcess(c *config.Config) bool {
	for _, cd := range c.Devices {
		if len(cd.Lvm) > 0 {
			return true
		}
	}
	return false
}
