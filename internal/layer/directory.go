package layer

import (
	"fmt"
	"os"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type CreateDirectoryLayer struct {
	fileBackend backend.FileBackend
}

func NewCreateDirectoryLayer(fb backend.FileBackend) *CreateDirectoryLayer {
	return &CreateDirectoryLayer{
		fileBackend: fb,
	}
}

func (fdl *CreateDirectoryLayer) From(c *config.Config) error {
	return fdl.fileBackend.From(c)
}

func (fdl *CreateDirectoryLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}

		d, err := fdl.fileBackend.GetDirectory(cd.MountPoint)
		if err != nil && !os.IsNotExist(err) {
			// This layer's responsibility is to create a directory if it doesn't exist.
			// Therefore, it won't generate an error if a file doesn't exist at the specified path.
			return nil, fmt.Errorf("🔴 %s: %s must be a directory for a device to be mounted to it", name, cd.MountPoint)
		}
		if d != nil {
			continue
		}

		mode := c.GetMode(name)
		a := fdl.fileBackend.CreateDirectory(cd.MountPoint).SetMode(mode)
		actions = append(actions, a)
	}
	return actions, nil
}

func (fdl *CreateDirectoryLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}
		if _, err := fdl.fileBackend.GetDirectory(cd.MountPoint); err != nil {
			return fmt.Errorf("🔴 %s: Failed directory validation checks. %s does not exist or is not a directory", name, cd.MountPoint)
		}
	}
	return nil
}

func (fdl *CreateDirectoryLayer) Warning() string {
	return DisabledWarning
}

func (fdl *CreateDirectoryLayer) ShouldProcess(c *config.Config) bool {
	for _, cd := range c.Devices {
		if len(cd.MountPoint) > 0 {
			return true
		}
	}
	return false
}
