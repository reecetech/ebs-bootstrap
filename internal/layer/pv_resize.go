package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

const (
	// The % threshold at which to resize a physical volume
	// -------------------------------------------------------
	// If the (physical volume / device size) * 100 falls
	// under this threshold then we perform a resize operation
	// -------------------------------------------------------
	// The smallest gp3 EBS volume you can create is 1GiB (1073741824 bytes).
	// The default size of the extent of a PV is 4 MiB (4194304 bytes).
	// Typically, the first extent of a PV is reserved for metadata. This
	// produces a PV of size 1069547520 bytes (Usage=99.6093%). We ensure
	// that we set the resize threshold to 99.6% to ensure that a 1 GiB EBS
	// volume won't be always resized
	// -------------------------------------------------------
	// Why not just look for a difference of 4194304 bytes?
	//	- The size of the extent can be changed by the user
	//	- Therefore we may not always see a difference of 4194304 bytes between
	//	  the block device and physical volume size
	ResizeThreshold = float64(99.6)
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
		if !c.GetResizeFs(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizePhysicalVolume(name, ResizeThreshold)
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
		if !c.GetResizeFs(name) {
			continue
		}
		shouldResize, err := rpvl.lvmBackend.ShouldResizePhysicalVolume(name, ResizeThreshold)
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
		if len(cd.Lvm) > 0 && c.GetResizeFs(name) {
			return true
		}
	}
	return false
}
