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

func TestCreateDirectoryLayerModify(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Files         map[string]*model.File
		CmpOption     cmp.Option
		ExpectedOuput []action.Action
		ExpectedError error
	}{
		{
			Name: "Mount Point Is a Directory",
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
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Mount Point Does Not Exist",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			Files:     map[string]*model.File{},
			CmpOption: cmp.AllowUnexported(action.CreateDirectoryAction{}),
			ExpectedOuput: []action.Action{
				action.NewCreateDirectoryAction("/mnt/foo", nil).SetMode(config.DefaultMode),
			},
			ExpectedError: nil,
		},
		{
			Name: "Mount Point is not a Directory",
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
					Type: model.RegularFile,
				},
			},
			CmpOption:     cmp.AllowUnexported(action.CreateDirectoryAction{}),
			ExpectedOuput: nil,
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: /mnt/foo must be a directory for a device to be mounted to it"),
		},
		{
			Name: "Ignore Mount Point Creation",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			Files:         map[string]*model.File{},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
		{
			Name: "Mount Point Is a Symbolic Link",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/bar",
					},
				},
			},
			Files: map[string]*model.File{
				"/mnt/bar": {
					Path: "/mnt/foo",
					Type: model.Directory,
				},
			},
			CmpOption:     cmp.AllowUnexported(),
			ExpectedOuput: []action.Action{},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			cdl := NewCreateDirectoryLayer(lfb)
			actions, err := cdl.Modify(subtest.Config)
			utils.CheckError("cdl.Modify()", t, subtest.ExpectedError, err)
			utils.CheckOutput("cdl.Modify()", t, subtest.ExpectedOuput, actions, subtest.CmpOption)
		})
	}
}

func TestCreateDirectoryLayerValidate(t *testing.T) {
	subtests := []struct {
		Name          string
		Config        *config.Config
		Files         map[string]*model.File
		ExpectedError error
	}{
		{
			Name: "Mount Point Exists As Directory",
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
		{
			Name: "Mount Point Does Not Exist",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			Files:         map[string]*model.File{},
			ExpectedError: fmt.Errorf("🔴 /dev/xvdf: Failed directory validation checks. /mnt/foo does not exist or is not a directory"),
		},
		{
			Name: "Skip Mount Point Validation",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			Files:         map[string]*model.File{},
			ExpectedError: nil,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := backend.NewMockLinuxFileBackend(subtest.Files)
			cdl := NewCreateDirectoryLayer(lfb)
			err := cdl.Validate(subtest.Config)
			utils.CheckError("cdl.Validate()", t, subtest.ExpectedError, err)
		})
	}
}

func TestCreateDirectoryLayerShouldProcess(t *testing.T) {
	subtests := []struct {
		Name           string
		Config         *config.Config
		ExpectedOutput bool
	}{
		{
			Name: "At Least Once Device Has Mount Point Specified",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdb": {},
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			ExpectedOutput: true,
		},
		{
			Name: "No Device Has Mount Point Specified",
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
			cdl := NewCreateDirectoryLayer(nil)
			value := cdl.ShouldProcess(subtest.Config)
			utils.CheckOutput("cdl.ShouldProcess()", t, subtest.ExpectedOutput, value)
		})
	}
}
