package backend

import (
	"fmt"
	"testing"

	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestGetUser(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		User           string
		Users          map[string]*model.User
		ExpectedOutput *model.User
		ExpectedError  error
	}{
		{
			Name: "Valid User",
			User: "example",
			Users: map[string]*model.User{
				"example": {
					Name: "example",
					Id:   1000,
				},
			},
			ExpectedOutput: &model.User{
				Name: "example",
				Id:   1000,
			},
			ExpectedError: nil,
		},
		{
			Name:           "Non-existent User",
			User:           "does-not-exist",
			Users:          map[string]*model.User{},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 User does-not-exist does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxOwnerBackend(subtest.Users, nil)
			bd, err := ldb.GetUser(subtest.User)
			utils.CheckError("ldb.GetUser()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ldb.GetUser()", t, subtest.ExpectedOutput, bd)
		})
	}
}

func TestGetGroup(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		Group          string
		Groups         map[string]*model.Group
		ExpectedOutput *model.Group
		ExpectedError  error
	}{
		{
			Name:  "Valid Group",
			Group: "example",
			Groups: map[string]*model.Group{
				"example": {
					Name: "example",
					Id:   2000,
				},
			},
			ExpectedOutput: &model.Group{
				Name: "example",
				Id:   2000,
			},
			ExpectedError: nil,
		},
		{
			Name:           "Non-existent Group",
			Group:          "does-not-exist",
			Groups:         map[string]*model.Group{},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 Group does-not-exist does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			ldb := NewMockLinuxOwnerBackend(nil, subtest.Groups)
			bd, err := ldb.GetGroup(subtest.Group)
			utils.CheckError("ldb.GetGroup()", t, subtest.ExpectedError, err)
			utils.CheckOutput("ldb.GetGroup()", t, subtest.ExpectedOutput, bd)
		})
	}
}

func TestLinuxOwnerBackendFrom(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		GetUser        func(user string) (*model.User, error)
		GetGroup       func(group string) (*model.Group, error)
		ExpectedUsers  map[string]*model.User
		ExpectedGroups map[string]*model.Group
		ExpectedError  error
	}{
		{
			Name: "Valid User<string> and Group<int>",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						User:  "example",
						Group: "2000",
					},
				},
			},
			GetUser: func(user string) (*model.User, error) {
				return &model.User{
					Name: "example",
					Id:   1000,
				}, nil
			},
			GetGroup: func(group string) (*model.Group, error) {
				return &model.Group{
					Name: "example",
					Id:   2000,
				}, nil
			},
			ExpectedUsers: map[string]*model.User{
				"example": {
					Name: "example",
					Id:   1000,
				},
			},
			ExpectedGroups: map[string]*model.Group{
				"2000": {
					Name: "example",
					Id:   2000,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Valid User<int> and Invalid Group<string>",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						User:  "1000",
						Group: "does-not-exist",
					},
				},
			},
			GetUser: func(user string) (*model.User, error) {
				return &model.User{
					Name: "example",
					Id:   1000,
				}, nil
			},
			GetGroup: func(group string) (*model.Group, error) {
				return nil, fmt.Errorf("🔴 Group (name=%s) does not exist", group)
			},
			ExpectedUsers:  nil,
			ExpectedGroups: nil,
			ExpectedError:  fmt.Errorf("🔴 Group (name=does-not-exist) does not exist"),
		},
		{
			Name: "Invalid User<int> and Valid Group<string>",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						User:  "-1",
						Group: "example",
					},
				},
			},
			GetUser: func(user string) (*model.User, error) {
				return nil, fmt.Errorf("🔴 User (id=%s) does not exist", user)
			},
			GetGroup: func(group string) (*model.Group, error) {
				return &model.Group{
					Name: "example",
					Id:   2000,
				}, nil
			},
			ExpectedUsers:  nil,
			ExpectedGroups: nil,
			ExpectedError:  fmt.Errorf("🔴 User (id=-1) does not exist"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			mos := service.NewMockOwnerService()
			if subtest.GetUser != nil {
				mos.StubGetUser = subtest.GetUser
			}
			if subtest.GetGroup != nil {
				mos.StubGetGroup = subtest.GetGroup
			}

			lob := NewLinuxOwnerBackend(mos)
			err := lob.From(subtest.Config)
			utils.CheckError("lob.From()", t, subtest.ExpectedError, err)
			utils.CheckOutput("lob.From()", t, subtest.ExpectedUsers, lob.users)
			utils.CheckOutput("lob.From()", t, subtest.ExpectedGroups, lob.groups)
		})
	}
}
