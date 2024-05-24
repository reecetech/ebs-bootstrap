package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type ResizeDeviceLayer struct {
	deviceBackend        backend.DeviceBackend
	deviceMetricsBackend backend.DeviceMetricsBackend
}

func NewResizeDeviceLayer(db backend.DeviceBackend, dmb backend.DeviceMetricsBackend) *ResizeDeviceLayer {
	return &ResizeDeviceLayer{
		deviceBackend:        db,
		deviceMetricsBackend: dmb,
	}
}

func (fdl *ResizeDeviceLayer) From(c *config.Config) error {
	err := fdl.deviceBackend.From(c)
	if err != nil {
		return err
	}
	return fdl.deviceMetricsBackend.From(c)
}

func (fdl *ResizeDeviceLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name := range c.Devices {
		if !c.GetResize(name) {
			continue
		}

		bd, err := fdl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return nil, err
		}
		metrics, err := fdl.deviceMetricsBackend.GetBlockDeviceMetrics(name)
		if err != nil {
			return nil, err
		}

		if !fdl.deviceMetricsBackend.ShouldResize(metrics) {
			continue
		}
		mode := c.GetMode(name)
		a, err := fdl.deviceBackend.Resize(bd)
		if err != nil {
			return nil, err
		}
		a = a.SetMode(mode)
		actions = append(actions, a)
	}
	return actions, nil
}

func (fdl *ResizeDeviceLayer) Validate(c *config.Config) error {
	for name := range c.Devices {
		if !c.GetResize(name) {
			continue
		}
		metrics, err := fdl.deviceMetricsBackend.GetBlockDeviceMetrics(name)
		if err != nil {
			return err
		}
		if fdl.deviceMetricsBackend.ShouldResize(metrics) {
			return fmt.Errorf("🔴 %s: Failed to resize file system. File System=%d Block Device=%d (bytes)", name, metrics.FileSystemSize, metrics.BlockDeviceSize)
		}
	}
	return nil
}

func (fdl *ResizeDeviceLayer) Warning() string {
	return DisabledWarning
}

func (fdl *ResizeDeviceLayer) ShouldProcess(c *config.Config) bool {
	for name := range c.Devices {
		if c.GetResize(name) {
			return true
		}
	}
	return false
}
