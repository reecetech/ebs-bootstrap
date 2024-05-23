package backend

import (
	"fmt"

	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type OwnerBackend interface {
	GetUser(user string) (*model.User, error)
	GetGroup(group string) (*model.Group, error)
	From(config *config.Config) error
}

type LinuxOwnerBackend struct {
	users        map[string]*model.User
	groups       map[string]*model.Group
	ownerService service.OwnerService
}

func NewLinuxOwnerBackend(ows service.OwnerService) *LinuxOwnerBackend {
	return &LinuxOwnerBackend{
		users:        map[string]*model.User{},
		groups:       map[string]*model.Group{},
		ownerService: ows,
	}
}

func NewMockLinuxOwnerBackend(users map[string]*model.User, groups map[string]*model.Group) *LinuxOwnerBackend {
	return &LinuxOwnerBackend{
		users:        users,
		groups:       groups,
		ownerService: nil,
	}
}

func (lfb *LinuxOwnerBackend) GetUser(user string) (*model.User, error) {
	o, exists := lfb.users[user]
	if !exists {
		return nil, fmt.Errorf("🔴 User %s does not exist", user)
	}
	return o, nil
}

func (lfb *LinuxOwnerBackend) GetGroup(group string) (*model.Group, error) {
	g, exists := lfb.groups[group]
	if !exists {
		return nil, fmt.Errorf("🔴 Group %s does not exist", group)
	}
	return g, nil
}

func (lfb *LinuxOwnerBackend) From(config *config.Config) error {
	lfb.users = nil
	lfb.groups = nil
	users := map[string]*model.User{}
	groups := map[string]*model.Group{}

	for _, cd := range config.Devices {
		if len(cd.User) > 0 {
			o, err := lfb.ownerService.GetUser(cd.User)
			if err != nil {
				return err
			}
			users[cd.User] = o
		}
		if len(cd.Group) > 0 {
			g, err := lfb.ownerService.GetGroup(cd.Group)
			if err != nil {
				return err
			}
			groups[cd.Group] = g
		}
	}
	lfb.users = users
	lfb.groups = groups
	return nil
}
