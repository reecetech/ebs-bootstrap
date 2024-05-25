package layer

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/backend"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestChangeOwnerLayerModify(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		Users          map[string]*model.User
		Groups         map[string]*model.Group
		Files          map[string]*model.File
		CmpOption      cmp.Option
		ExpectedOutput []action.Action
		ExpectedError  error
	}{
		{
			Name: "Change the User of The Mount Point",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
						User:       "user-b",
						Group:      "group-a",
					},
				},
			},
			Users: map[string]*model.User{
				"user-b": {
					Name: "user-b",
					Id:   model.UserId(1500),
				},
			},
			Groups: map[string]*model.Group{
				"group-a": {
					Name: "group-a",
					Id:   model.GroupId(2000),
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:    "/mnt/foo",
					Type:    model.Directory,
					UserId:  model.UserId(1000),
					GroupId: model.GroupId(2000),
				},
			},
			CmpOption: cmp.AllowUnexported(action.ChangeOwnerAction{}),
			ExpectedOutput: []action.Action{
				action.NewChangeOwnerAction("/mnt/foo", model.UserId(1500), model.GroupId(2000), nil).SetMode(config.DefaultMode),
			},
		},
		{
			Name: "Skip + Mount Point Not Provided",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			Users:          map[string]*model.User{},
			Groups:         map[string]*model.Group{},
			Files:          map[string]*model.File{},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: []action.Action{},
			ExpectedError:  nil,
		},
		{
			Name: "Invalid + Mount Point Does Not Exist",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
						User:       "user-a",
						Group:      "group-a",
					},
				},
			},
			Users: map[string]*model.User{
				"user-a": {
					Name: "user-a",
					Id:   model.UserId(1000),
				},
			},
			Groups: map[string]*model.Group{
				"group-a": {
					Name: "group-a",
					Id:   model.GroupId(2000),
				},
			},
			Files:          map[string]*model.File{},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /mnt/foo is either not a directory or does not exist"),
		},
		{
			Name: "Skip + User and Group Not Provided",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			Users:  map[string]*model.User{},
			Groups: map[string]*model.Group{},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path: "/mnt/foo",
				},
			},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: []action.Action{},
			ExpectedError:  nil,
		},
		{
			Name: "Skip + User and Group Do Not Change",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
						User:       "user-a",
						Group:      "group-a",
					},
				},
			},
			Users: map[string]*model.User{
				"user-a": {
					Name: "user-a",
					Id:   model.UserId(1000),
				},
			},
			Groups: map[string]*model.Group{
				"group-a": {
					Name: "group-a",
					Id:   model.GroupId(2000),
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:    "/mnt/foo",
					Type:    model.Directory,
					UserId:  model.UserId(1000),
					GroupId: model.GroupId(2000),
				},
			},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: []action.Action{},
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			lob := backend.NewMockLinuxOwnerBackend(subtest.Users, subtest.Groups)
			col := NewChangeOwnerLayer(lob, lfb)
			actions, err := col.Modify(subtest.Config)
			utils.CheckError("col.Modify()", t, subtest.ExpectedError, err)
			utils.CheckOutput("col.Modify()", t, subtest.ExpectedOutput, actions, subtest.CmpOption)
		})
	}
}

func TestChangeOwnerLayerValidate(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Users         map[string]*model.User
		Groups        map[string]*model.Group
		Files         map[string]*model.File
		ExpectedError error
	}{
		{
			Name: "User and Group Match Requested User and Group",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
						User:       "user-a",
						Group:      "group-a",
					},
				},
			},
			Users: map[string]*model.User{
				"user-a": {
					Name: "user-a",
					Id:   model.UserId(1000),
				},
			},
			Groups: map[string]*model.Group{
				"group-a": {
					Name: "group-a",
					Id:   model.GroupId(2000),
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:    "/mnt/foo",
					Type:    model.Directory,
					UserId:  model.UserId(1000),
					GroupId: model.GroupId(2000),
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid + Mount Point Does Not Exist",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
						User:       "user-a",
						Group:      "group-a",
					},
				},
			},
			Users: map[string]*model.User{
				"user-a": {
					Name: "user-a",
					Id:   model.UserId(1000),
				},
			},
			Groups: map[string]*model.Group{
				"group-a": {
					Name: "group-a",
					Id:   model.GroupId(2000),
				},
			},
			Files:         map[string]*model.File{},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed ownership validation checks. /mnt/foo is either not a directory or does not exist"),
		},
		{
			Name: "Invalid + User Does Not Match Requested User",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
						User:       "user-b",
						Group:      "group-a",
					},
				},
			},
			Users: map[string]*model.User{
				"user-b": {
					Name: "user-b",
					Id:   model.UserId(1500),
				},
			},
			Groups: map[string]*model.Group{
				"group-a": {
					Name: "group-a",
					Id:   model.GroupId(2000),
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:    "/mnt/foo",
					Type:    model.Directory,
					UserId:  model.UserId(1000),
					GroupId: model.GroupId(2000),
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed ownership validation checks. /mnt/foo User Expected=1500, Actual=1000"),
		},
		{
			Name: "Invalid + Group Does Not Match Requested Group",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
						User:       "user-a",
						Group:      "group-b",
					},
				},
			},
			Users: map[string]*model.User{
				"user-a": {
					Name: "user-a",
					Id:   model.UserId(1000),
				},
			},
			Groups: map[string]*model.Group{
				"group-b": {
					Name: "group-b",
					Id:   model.GroupId(2500),
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:    "/mnt/foo",
					Type:    model.Directory,
					UserId:  model.UserId(1000),
					GroupId: model.GroupId(2000),
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed ownership validation checks. /mnt/foo Group Expected=2500, Actual=2000"),
		},
		{
			Name: "Skip + Mount Point Not Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			Users:         map[string]*model.User{},
			Groups:        map[string]*model.Group{},
			Files:         map[string]*model.File{},
			ExpectedError: nil,
		},
		{
			Name: "Skip + User and Group Not Provided",
			Config: &config.Config{

				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			Users:  map[string]*model.User{},
			Groups: map[string]*model.Group{},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path: "/mnt/foo",
				},
			},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			lob := backend.NewMockLinuxOwnerBackend(subtest.Users, subtest.Groups)
			col := NewChangeOwnerLayer(lob, lfb)
			err := col.Validate(subtest.Config)
			utils.CheckError("col.Validate()", t, subtest.ExpectedError, err)
		})
	}
}

func TestChangeOwnerLayerShouldProcess(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		ExpectedOutput bool
	}{
		{
			Name: "At Least One Device has Mount Point, User and Group Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						MountPoint: "/mnt/foo",
						User:       "user-a",
						Group:      "group-a",
					},
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: true,
		},
		{
			Name: "Device has Mount Point and User Provided, but not Group",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						MountPoint: "/mnt/foo",
						User:       "user-a",
					},
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: true,
		},
		{
			Name: "Device has Mount Point and Group Provided, but not User",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						MountPoint: "/mnt/foo",
						Group:      "group-a",
					},
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: true,
		},
		{
			Name: "Device has Mount Point Provided, but not User and Group",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {
						MountPoint: "/mnt/foo",
					},
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: false,
		},
		{
			Name: "No Device has Mount Point Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: false,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			col := NewChangeOwnerLayer(nil, nil)
			output := col.ShouldProcess(subtest.Config)
			utils.CheckOutput("col.ShouldProcess()", t, subtest.ExpectedOutput, output)
		})
	}
}
