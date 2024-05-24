package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

const (
	// The % tolerance to expect the logical volume size to be within
	// -------------------------------------------------------
	// If the (logical volume / volume group size) * 100 is less than
	// (lvmConsumption% - tolerance%) then we perform a resize operation
	// -------------------------------------------------------
	// If the (logical volume / volume group size) * 100 is greater than
	// (lvmConsumption% + tolerance%) then the user is attempting a downsize
	// operation. We outright deny this as downsizing can be a destructive
	// operation
	// -------------------------------------------------------
	// Why implement a tolernace-based policy for resizing?
	// 	- When creating a Logical Volume, `ebs-bootstrap` issues a command like
	// 		`lvcreate -l 20%VG -n lv_name vg_name`
	// 	- When we calculate how much percentage of the volume group has been
	// 		consumed by the logical volume, the value would look like 20.0052096...
	// 	- A tolerance establishes a window of acceptable values for avoiding a
	// 		resizing operation
	ResizeTolerance = float64(0.1)
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
		if !c.GetResizeFs(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizeLogicalVolume(cd.Lvm, cd.Lvm, c.GetLvmConsumption(name), ResizeTolerance)
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
		if !c.GetResizeFs(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizeLogicalVolume(cd.Lvm, cd.Lvm, c.GetLvmConsumption(name), ResizeTolerance)
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
		if len(cd.Lvm) > 0 && c.GetResizeFs(name) {
			return true
		}
	}
	return false
}
