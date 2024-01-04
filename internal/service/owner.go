package service

import (
	"fmt"
	"os/user"
	"strconv"

	"github.com/reecetech/ebs-bootstrap/internal/model"
)

type OwnerService interface {
	GetCurrentUser() (*model.User, error)
	GetCurrentGroup() (*model.Group, error)
	GetUser(usr string) (*model.User, error)
	GetGroup(grp string) (*model.Group, error)
}

type UnixOwnerService struct{}

func NewUnixOwnerService() *UnixOwnerService {
	return &UnixOwnerService{}
}

func (uos *UnixOwnerService) GetCurrentUser() (*model.User, error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("🔴 Could not get current user")
	}
	return uos.GetUser(u.Uid)
}

func (uos *UnixOwnerService) GetCurrentGroup() (*model.Group, error) {
	u, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("🔴 Could not get current user")
	}
	return uos.GetGroup(u.Gid)
}

func (uos *UnixOwnerService) GetUser(usr string) (*model.User, error) {
	var u *user.User
	if _, err := strconv.Atoi(usr); err != nil {
		// If not a valid integer, try to look up by username
		u, err = user.Lookup(usr)
		if err != nil {
			return nil, fmt.Errorf("🔴 User (name=%s) does not exist", usr)
		}
	} else {
		u, err = user.LookupId(usr)
		if err != nil {
			return nil, fmt.Errorf("🔴 User (id=%s) does not exist", usr)
		}
	}
	uid, err := strconv.ParseUint(u.Uid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to cast user id to unsigned 32-bit integer")
	}
	return &model.User{Name: u.Username, Id: model.UserId(uid)}, nil
}

func (uos *UnixOwnerService) GetGroup(grp string) (*model.Group, error) {
	var g *user.Group
	if _, err := strconv.Atoi(grp); err != nil {
		// If not a valid integer, try to look up by group name
		g, err = user.LookupGroup(grp)
		if err != nil {
			return nil, fmt.Errorf("🔴 Group (name=%s) does not exist", grp)
		}
	} else {
		g, err = user.LookupGroupId(grp)
		if err != nil {
			return nil, fmt.Errorf("🔴 Group (id=%s) does not exist", grp)
		}
	}
	gid, err := strconv.ParseUint(g.Gid, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to cast group id to unsigned 32-bit integer")
	}
	return &model.Group{Name: g.Name, Id: model.GroupId(gid)}, nil
}
