package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type ResizeLogicalVolumeLayer struct {
	lvmBackend backend.LvmBackend
}

func NewResizeLogicalVolumeLayer(lb backend.LvmBackend) *ResizeLogicalVolumeLayer {
	return &ResizeLogicalVolumeLayer{
		lvmBackend: lb,
	}
}

func (rpvl *ResizeLogicalVolumeLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		if !c.GetResize(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizeLogicalVolume(cd.Lvm, cd.Lvm, c.GetLvmConsumption(name))
		if err != nil {
			return nil, err
		}
		if !shouldResize {
			continue
		}
		mode := c.GetMode(name)
		a := rpvl.lvmBackend.ResizeLogicalVolume(cd.Lvm, cd.Lvm, c.GetLvmConsumption(name))
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (rpvl *ResizeLogicalVolumeLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		if !c.GetResize(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizeLogicalVolume(cd.Lvm, cd.Lvm, c.GetLvmConsumption(name))
		if err != nil {
			return err
		}
		if shouldResize {
			return fmt.Errorf("🔴 %s: Failed resize validation checks. Logical volume %s still needs to be resized", name, name)
		}
	}
	return nil
}

func (rpvl *ResizeLogicalVolumeLayer) Warning() string {
	return DisabledWarning
}

func (rpvl *ResizeLogicalVolumeLayer) From(c *config.Config) error {
	return rpvl.lvmBackend.From(c)
}

func (rpvl *ResizeLogicalVolumeLayer) ShouldProcess(c *config.Config) bool {
	for name, cd := range c.Devices {
		if len(cd.Lvm) > 0 && c.GetResize(name) {
			return true
		}
	}
	return false
}
