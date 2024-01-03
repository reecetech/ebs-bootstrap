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

func TestChangePermissionsLayerModify(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		Files          map[string]*model.File
		CmpOption      cmp.Option
		ExpectedOutput []action.Action
		ExpectedError  error
	}{
		{
			Name: "Change the Permissions of The Mount Point",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint:  "/mnt/foo",
						Permissions: model.FilePermissions(0755),
					},
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:        "/mnt/foo",
					Type:        model.Directory,
					Permissions: model.FilePermissions(0644),
				},
			},
			CmpOption: cmp.AllowUnexported(action.ChangePermissionsAction{}),
			ExpectedOutput: []action.Action{
				action.NewChangePermissionsAction("/mnt/foo", model.FilePermissions(0755), nil).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid + Mount Point Does Not Exist",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint:  "/mnt/foo",
						Permissions: model.FilePermissions(0755),
					},
				},
			},
			Files:          map[string]*model.File{},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /mnt/foo is either not a directory or does not exist"),
		},
		{
			Name: "Mount Point Permissions Match Requested Permissions",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint:  "/mnt/foo",
						Permissions: model.FilePermissions(0755),
					},
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:        "/mnt/foo",
					Type:        model.Directory,
					Permissions: model.FilePermissions(0755),
				},
			},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: []action.Action{},
			ExpectedError:  nil,
		},
		{
			Name: "Skip + No Mount Point",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			Files:          map[string]*model.File{},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: []action.Action{},
			ExpectedError:  nil,
		},
		{
			Name: "Skip + No Permissions Declared",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path: "/mnt/foo",
					Type: model.Directory,
				},
			},
			CmpOption:      cmp.AllowUnexported(),
			ExpectedOutput: []action.Action{},
			ExpectedError:  nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			cpl := NewChangePermissionsLayer(lfb)
			actions, err := cpl.Modify(subtest.Config)
			utils.CheckError("cpl.Modify()", t, subtest.ExpectedError, err)
			utils.CheckOutput("cpl.Modify()", t, subtest.ExpectedOutput, actions, subtest.CmpOption)
		})
	}
}

func TestChangePermissionsLayerValidate(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Files         map[string]*model.File
		ExpectedError error
	}{
		{
			Name: "Mount Point Permissions Match Requested Permissions",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint:  "/mnt/foo",
						Permissions: model.FilePermissions(0755),
					},
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:        "/mnt/foo",
					Type:        model.Directory,
					Permissions: model.FilePermissions(0755),
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid + Mount Point Does Not Exist",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint:  "/mnt/foo",
						Permissions: model.FilePermissions(0755),
					},
				},
			},
			Files:         map[string]*model.File{},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed ownership validation checks. /mnt/foo is either not a directory or does not exist"),
		},
		{
			Name: "Invalid + Mount Point Permissions Do Not Match Requested Permissions",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint:  "/mnt/foo",
						Permissions: model.FilePermissions(0755),
					},
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:        "/mnt/foo",
					Type:        model.Directory,
					Permissions: model.FilePermissions(0644),
				},
			},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed permissions validation checks. /mnt/foo Permissions Expected=0755, Actual=0644"),
		},
		{
			Name: "Skip + No Mount Point",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			Files:         map[string]*model.File{},
			ExpectedError: nil,
		},
		{
			Name: "Skip + No Permissions Declared",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path: "/mnt/foo",
					Type: model.Directory,
				},
			},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			cpl := NewChangePermissionsLayer(lfb)
			err := cpl.Validate(subtest.Config)
			utils.CheckError("cpl.Modify()", t, subtest.ExpectedError, err)
		})
	}
}
