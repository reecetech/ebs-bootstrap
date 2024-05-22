package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
)

type ActivateLogicalVolumeLayer struct {
	lvmBackend backend.LvmBackend
}

func NewActivateLogicalVolumeLayer(lb backend.LvmBackend) *ActivateLogicalVolumeLayer {
	return &ActivateLogicalVolumeLayer{
		lvmBackend: lb,
	}
}

func (cvgl *ActivateLogicalVolumeLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}

		lv, err := cvgl.lvmBackend.GetLogicalVolume(cd.Lvm, cd.Lvm)
		if err != nil {
			return nil, err
		}

		if lv.State == model.Active {
			continue
		}

		if lv.State == model.Unsupported {
			return nil, fmt.Errorf("🔴 %s: Can not activate a logical volume in an unsupported state", lv.Name)
		}

		mode := c.GetMode(name)
		a := cvgl.lvmBackend.ActivateLogicalVolume(cd.Lvm, cd.Lvm)
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (cvgl *ActivateLogicalVolumeLayer) Validate(c *config.Config) error {
	for _, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
	}
	return nil
}

func (cvgl *ActivateLogicalVolumeLayer) Warning() string {
	return DisabledWarning
}

func (cvgl *ActivateLogicalVolumeLayer) From(c *config.Config) error {
	return cvgl.lvmBackend.From(c)
}

func (cvgl *ActivateLogicalVolumeLayer) ShouldProcess(c *config.Config) bool {
	for _, cd := range c.Devices {
		if len(cd.Lvm) > 0 {
			return true
		}
	}
	return false
}
