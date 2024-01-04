package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type LabelDeviceLayer struct {
	deviceBackend backend.DeviceBackend
}

func NewLabelDeviceLayer(db backend.DeviceBackend) *LabelDeviceLayer {
	return &LabelDeviceLayer{
		deviceBackend: db,
	}
}

func (fdl *LabelDeviceLayer) From(c *config.Config) error {
	return fdl.deviceBackend.From(c)
}

func (fdl *LabelDeviceLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.Label) == 0 {
			continue
		}

		bd, err := fdl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return nil, err
		}
		if bd.Label == cd.Label {
			continue
		}

		mode := c.GetMode(name)
		las, err := fdl.deviceBackend.Label(bd, cd.Label)
		if err != nil {
			return nil, err
		}
		for _, la := range las {
			actions = append(actions, la.SetMode(mode))
		}
	}
	return actions, nil
}

func (fdl *LabelDeviceLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.Label) == 0 {
			continue
		}
		d, err := fdl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return err
		}
		if d.Label != cd.Label {
			return fmt.Errorf("🔴 %s: Failed label validation checks. Expected=%s, Actual=%s", name, cd.Label, d.Label)
		}
	}
	return nil
}

func (fdl *LabelDeviceLayer) Warning() string {
	return "Certain file systems require that devices be unmounted prior to labeling"
}
