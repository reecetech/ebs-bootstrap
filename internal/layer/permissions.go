package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type ChangePermissionsLayer struct {
	fileBackend backend.FileBackend
}

func NewChangePermissionsLayer(fb backend.FileBackend) *ChangePermissionsLayer {
	return &ChangePermissionsLayer{
		fileBackend: fb,
	}
}

func (fdl *ChangePermissionsLayer) From(c *config.Config) error {
	return fdl.fileBackend.From(c)
}

func (fdl *ChangePermissionsLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}
		if cd.Permissions == 0 {
			continue
		}

		d, err := fdl.fileBackend.GetDirectory(cd.MountPoint)
		if err != nil {
			return nil, fmt.Errorf("🔴 %s is either not a directory or does not exist", cd.MountPoint)
		}

		if d.Permissions == cd.Permissions {
			continue
		}

		mode := c.GetMode(name)
		a := fdl.fileBackend.ChangePermissions(cd.MountPoint, cd.Permissions).SetMode(mode)
		actions = append(actions, a)
	}
	return actions, nil
}

func (fdl *ChangePermissionsLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}
		if cd.Permissions == 0 {
			continue
		}

		d, err := fdl.fileBackend.GetDirectory(cd.MountPoint)
		if err != nil {
			return fmt.Errorf("🔴 %s: Failed ownership validation checks. %s is either not a directory or does not exist", name, cd.MountPoint)
		}

		if d.Permissions != cd.Permissions {
			return fmt.Errorf("🔴 %s: Failed permissions validation checks. %s Permissions Expected=%#o, Actual=%#o", name, cd.MountPoint, cd.Permissions, d.Permissions)
		}
	}
	return nil
}

func (fdl *ChangePermissionsLayer) Warning() string {
	return DisabledWarning
}

func (fdl *ChangePermissionsLayer) ShouldProcess(c *config.Config) bool {
	for _, cd := range c.Devices {
		if len(cd.MountPoint) > 0 && cd.Permissions != 0 {
			return true
		}
	}
	return false
}
