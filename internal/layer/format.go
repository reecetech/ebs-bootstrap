package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
)

type FormatDeviceLayer struct {
	deviceBackend backend.DeviceBackend
}

func NewFormatDeviceLayer(db backend.DeviceBackend) *FormatDeviceLayer {
	return &FormatDeviceLayer{
		deviceBackend: db,
	}
}

func (fdl *FormatDeviceLayer) From(c *config.Config) error {
	return fdl.deviceBackend.From(c)
}

func (fdl *FormatDeviceLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if cd.Fs == model.Unformatted {
			return nil, fmt.Errorf("🔴 %s: Can not erase the file system of a device", name)
		}

		bd, err := fdl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return nil, err
		}
		if bd.FileSystem == cd.Fs {
			continue
		}
		if bd.FileSystem != model.Unformatted {
			return nil, fmt.Errorf("🔴 %s: Can not format a device with an existing %s file system", bd.Name, bd.FileSystem.String())
		}

		mode := c.GetMode(name)
		a, err := fdl.deviceBackend.Format(bd, cd.Fs)
		if err != nil {
			return nil, err
		}
		actions = append(actions, a.SetMode(mode))
	}
	return actions, nil
}

func (fdl *FormatDeviceLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		d, err := fdl.deviceBackend.GetBlockDevice(name)
		if err != nil {
			return err
		}
		if d.FileSystem != cd.Fs {
			return fmt.Errorf("🔴 %s: Failed file system validation checks. Expected=%s, Actual=%s", name, cd.Fs, d.FileSystem.String())
		}
	}
	return nil
}

func (fdl *FormatDeviceLayer) Warning() string {
	return "Formatting larger disks can take several seconds ⌛"
}

func (fdl *FormatDeviceLayer) ShouldProcess(c *config.Config) bool {
	return true
}
