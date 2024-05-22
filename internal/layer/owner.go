package layer

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
)

type ChangeOwnerLayer struct {
	ownerBackend backend.OwnerBackend
	fileBackend  backend.FileBackend
}

func NewChangeOwnerLayer(ub backend.OwnerBackend, fb backend.FileBackend) *ChangeOwnerLayer {
	return &ChangeOwnerLayer{
		ownerBackend: ub,
		fileBackend:  fb,
	}
}

func (fdl *ChangeOwnerLayer) From(c *config.Config) error {
	err := fdl.ownerBackend.From(c)
	if err != nil {
		return err
	}
	return fdl.fileBackend.From(c)
}

func (fdl *ChangeOwnerLayer) Modify(c *config.Config) ([]action.Action, error) {
	actions := make([]action.Action, 0)
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}
		if len(cd.User) == 0 && len(cd.Group) == 0 {
			continue
		}

		d, err := fdl.fileBackend.GetDirectory(cd.MountPoint)
		if err != nil {
			return nil, fmt.Errorf("🔴 %s is either not a directory or does not exist", cd.MountPoint)
		}
		uid := d.UserId
		gid := d.GroupId
		if len(cd.User) > 0 {
			u, err := fdl.ownerBackend.GetUser(cd.User)
			if err != nil {
				return nil, err
			}
			uid = u.Id
		}

		if len(cd.Group) > 0 {
			g, err := fdl.ownerBackend.GetGroup(cd.Group)
			if err != nil {
				return nil, err
			}
			gid = g.Id
		}
		if d.UserId == uid && d.GroupId == gid {
			continue
		}

		mode := c.GetMode(name)
		a := fdl.fileBackend.ChangeOwner(cd.MountPoint, uid, gid).SetMode(mode)
		actions = append(actions, a)
	}
	return actions, nil
}

func (fdl *ChangeOwnerLayer) Validate(c *config.Config) error {
	for name, cd := range c.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}
		if len(cd.User) == 0 && len(cd.Group) == 0 {
			continue
		}

		d, err := fdl.fileBackend.GetDirectory(cd.MountPoint)
		if err != nil {
			return fmt.Errorf("🔴 %s: Failed ownership validation checks. %s is either not a directory or does not exist", name, cd.MountPoint)
		}

		uid := d.UserId
		gid := d.GroupId
		if len(cd.User) > 0 {
			u, err := fdl.ownerBackend.GetUser(cd.User)
			if err != nil {
				return err
			}
			uid = u.Id
		}

		if len(cd.Group) > 0 {
			g, err := fdl.ownerBackend.GetGroup(cd.Group)
			if err != nil {
				return err
			}
			gid = g.Id
		}

		if d.UserId != uid {
			return fmt.Errorf("🔴 %s: Failed ownership validation checks. %s User Expected=%d, Actual=%d", name, cd.MountPoint, uid, d.UserId)
		}
		if d.GroupId != gid {
			return fmt.Errorf("🔴 %s: Failed ownership validation checks. %s Group Expected=%d, Actual=%d", name, cd.MountPoint, gid, d.GroupId)
		}
	}
	return nil
}

func (fdl *ChangeOwnerLayer) Warning() string {
	return DisabledWarning
}

func (fdl *ChangeOwnerLayer) ShouldProcess(c *config.Config) bool {
	for _, cd := range c.Devices {
		if len(cd.MountPoint) > 0 && (len(cd.User) > 0 || len(cd.Group) > 0) {
			return true
		}
	}
	return false
}
