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

		vg, _ := cvgl.lvmBackend.GetVolumeGroup(cd.Lvm)

		if vg != nil {
			pvs, err := cvgl.lvmBackend.SearchPhysicalVolumes(vg)
			if err != nil {
				return nil, err
			}
			if len(pvs) == 1 {
				if pvs[0].Name == name {
					continue
				}
				return nil, fmt.Errorf("🔴 %s: Volume group %s belongs to the incorrect physical volume. Expected=%s, Actual=%s", name, vg.Name, name, pvs[0].Name)
			}
			return nil, fmt.Errorf("🔴 %s: Cannot manage a volume group %s with more than one physical volumes associated", name, vg.Name)
		}

		pv, err := cvgl.lvmBackend.GetPhysicalVolume(name)
		if err != nil {
			return nil, err
		}

		vg, err = cvgl.lvmBackend.SearchVolumeGroup(pv)
		if err != nil {
			return nil, err
		}

		if vg != nil {
			return nil, fmt.Errorf("🔴 %s: Volume group %s already exists on physical volume %s", name, vg.Name, pv.Name)
		}

		mode := c.GetMode(name)
		a := cvgl.lvmBackend.CreateVolumeGroup(cd.Lvm, pv.Name)
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (cvgl *CreateVolumeGroupLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		_, err := cvgl.lvmBackend.GetVolumeGroup(cd.Lvm)
		if err != nil {
			return fmt.Errorf("🔴 %s: Failed volume group validation checks. Volume group %s does not exist", name, cd.Lvm)
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
