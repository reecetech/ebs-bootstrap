package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type ResizePhysicalVolumeLayer struct {
	lvmBackend backend.LvmBackend
}

func NewResizePhysicalVolumeLayer(lb backend.LvmBackend) *ResizePhysicalVolumeLayer {
	return &ResizePhysicalVolumeLayer{
		lvmBackend: lb,
	}
}

func (rpvl *ResizePhysicalVolumeLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		if !c.GetResize(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizePhysicalVolume(name)
		if err != nil {
			return nil, err
		}
		if !shouldResize {
			continue
		}
		mode := c.GetMode(name)
		a := rpvl.lvmBackend.ResizePhysicalVolume(name)
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (rpvl *ResizePhysicalVolumeLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.Lvm) == 0 {
			continue
		}
		if !c.GetResize(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizePhysicalVolume(name)
		if err != nil {
			return err
		}
		if shouldResize {
			return fmt.Errorf("🔴 %s: Failed resize validation checks. Physical volume %s still needs to be resized", name, name)
		}
	}
	return nil
}

func (rpvl *ResizePhysicalVolumeLayer) Warning() string {
	return DisabledWarning
}

func (rpvl *ResizePhysicalVolumeLayer) From(c *config.Config) error {
	return rpvl.lvmBackend.From(c)
}

func (rpvl *ResizePhysicalVolumeLayer) ShouldProcess(c *config.Config) bool {
	for name, cd := range c.Devices {
		if len(cd.Lvm) > 0 && c.GetResize(name) {
			return true
		}
	}
	return false
}
