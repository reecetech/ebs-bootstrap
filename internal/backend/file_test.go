package backend

import (
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

func TestCreateDirectory(t *testing.T) {
	lfb := NewMockLinuxFileBackend(nil)
	subtests := []struct {
		Name           string
		Path           string
		ExpectedOutput action.Action
	}{
		{
			Name:           "Valid Action",
			Path:           "/mnt/app",
			ExpectedOutput: action.NewCreateDirectoryAction("/mnt/app", nil),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			a := lfb.CreateDirectory(subtest.Path)
			utils.CheckOutput("lfb.CreateDirectory()", t, subtest.ExpectedOutput, a, cmp.AllowUnexported(action.CreateDirectoryAction{}))
		})
	}
}

func TestChangeOwner(t *testing.T) {
	lfb := NewMockLinuxFileBackend(nil)
	subtests := []struct {
		Name           string
		Path           string
		Uid            model.UserId
		Gid            model.GroupId
		ExpectedOutput action.Action
	}{
		{
			Name:           "Valid Action",
			Path:           "/mnt/app",
			Uid:            1000,
			Gid:            1000,
			ExpectedOutput: action.NewChangeOwnerAction("/mnt/app", 1000, 1000, nil),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			a := lfb.ChangeOwner(subtest.Path, subtest.Uid, subtest.Gid)
			utils.CheckOutput("lfb.ChangeOwner()", t, subtest.ExpectedOutput, a, cmp.AllowUnexported(action.ChangeOwnerAction{}))
		})
	}
}

func TestChangePermissions(t *testing.T) {
	lfb := NewMockLinuxFileBackend(nil)
	subtests := []struct {
		Name           string
		Path           string
		Perms          model.FilePermissions
		ExpectedOutput action.Action
	}{
		{
			Name:           "Valid Action",
			Path:           "/mnt/app",
			Perms:          0755,
			ExpectedOutput: action.NewChangePermissionsAction("/mnt/app", 0755, nil),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			a := lfb.ChangePermissions(subtest.Path, subtest.Perms)
			utils.CheckOutput("lfb.ChangePermissions()", t, subtest.ExpectedOutput, a, cmp.AllowUnexported(action.ChangePermissionsAction{}))
		})
	}
}

func TestGetDirectory(t *testing.T) {
	subtests := []struct {
		Name           string
		Files          map[string]*model.File
		Path           string
		ExpectedOutput *model.File
		ExpectedError  error
	}{
		{
			Name: "Valid Directory",
			Files: map[string]*model.File{
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
			},
			Path: "/mnt",
			ExpectedOutput: &model.File{
				Path: "/mnt",
				Type: model.Directory,
			},
			ExpectedError: nil,
		},
		{
			Name: "Invalid Directory",
			Files: map[string]*model.File{
				"/tmp/foo": {
					Path: "/tmp/foo",
					Type: model.RegularFile,
				},
			},
			Path:           "/tmp/foo",
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 /tmp/foo: File is not a directory"),
		},
		{
			Name: "Directory Does Not Exist",
			Files: map[string]*model.File{
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
			},
			Path:           "/mnt/foo",
			ExpectedOutput: nil,
			ExpectedError:  os.ErrNotExist,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := NewMockLinuxFileBackend(subtest.Files)
			f, err := lfb.GetDirectory(subtest.Path)
			utils.CheckError("lfb.GetDirectory()", t, subtest.ExpectedError, err)
			utils.CheckOutput("lfb.GetDirectory()", t, subtest.ExpectedOutput, f)
		})
	}
}

func TestIsMount(t *testing.T) {
	subtests := []struct {
		Name           string
		Files          map[string]*model.File
		Path           string
		ExpectedOutput bool
	}{
		{
			Name: "Child Path Does Not Exist",
			Files: map[string]*model.File{
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
			},
			Path:           "/mnt/foo",
			ExpectedOutput: false,
		},
		{
			Name: "Child Path Is Not Directory",
			Files: map[string]*model.File{
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
				"/mnt/foo": {
					Path: "/mnt/foo",
					Type: model.RegularFile,
				},
			},
			Path:           "/mnt/foo",
			ExpectedOutput: false,
		},
		// This is an unrealistic test case, but it is included for completeness.
		// It's not possible for a child file to exist if the parent directory does not exist
		{
			Name: "Child Directory Exists + Parent Directory Does Not Exist",
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path: "/mnt/foo",
					Type: model.Directory,
				},
			},
			Path:           "/mnt/foo",
			ExpectedOutput: false,
		},
		{
			Name: "Child Directory Is Mount",
			Files: map[string]*model.File{
				"/mnt/foo": {
					Path:     "/mnt/foo",
					Type:     model.Directory,
					DeviceId: 2000,
				},
				"/mnt": {
					Path:     "/mnt",
					Type:     model.Directory,
					DeviceId: 1000,
				},
			},
			Path:           "/mnt/foo",
			ExpectedOutput: true,
		},
		{
			Name: "Root Directory",
			Files: map[string]*model.File{
				"/": {
					Path:     "/",
					Type:     model.Directory,
					DeviceId: 1000,
					InodeNo:  2000,
				},
			},
			Path:           "/",
			ExpectedOutput: true,
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			lfb := NewMockLinuxFileBackend(subtest.Files)
			im := lfb.IsMount(subtest.Path)
			utils.CheckOutput("lfb.IsMount()", t, subtest.ExpectedOutput, im)
		})
	}
}

func TestLinuxFileBackendFrom(t *testing.T) {
	counter := 0
	subtests := []struct {
		Name           string
		Config         *config.Config
		GetFile        func(p string) (*model.File, error)
		ExpectedOutput map[string]*model.File
		ExpectedError  error
	}{
		{
			Name: "Mount Point (Exists) / Parent (Exists)",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			GetFile: func(p string) (*model.File, error) {
				return &model.File{
					Path: p,
					Type: model.Directory,
				}, nil
			},
			ExpectedOutput: map[string]*model.File{
				"/mnt/foo": {
					Path: "/mnt/foo",
					Type: model.Directory,
				},
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Mount Point (Does Not Exist) / Parent (Exists)",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			GetFile: func(p string) (*model.File, error) {
				switch p {
				case "/mnt/foo":
					return nil, os.ErrNotExist
				case "/mnt":
					return &model.File{Path: p, Type: model.Directory}, nil
				default:
					return nil, utils.NewNotImeplementedError("GetFile()")
				}
			},
			ExpectedOutput: map[string]*model.File{
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Triggering Deduplication Checks",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
					"/dev/xvdg": {
						MountPoint: "/mnt/bar",
					},
				},
			},
			GetFile: func(p string) (*model.File, error) {
				switch p {
				case "/mnt/foo":
					return &model.File{Path: p, Type: model.Directory}, nil
				case "/mnt/bar":
					return &model.File{Path: p, Type: model.Directory}, nil
				case "/mnt":
					if counter > 0 {
						return nil, fmt.Errorf(`🔴 GetFile("/mnt") should not be called more than once`)
					}
					counter++
					return &model.File{Path: p, Type: model.Directory}, nil
				default:
					return nil, utils.NewNotImeplementedError("GetFile()")
				}
			},
			ExpectedOutput: map[string]*model.File{
				"/mnt/foo": {
					Path: "/mnt/foo",
					Type: model.Directory,
				},
				"/mnt/bar": {
					Path: "/mnt/bar",
					Type: model.Directory,
				},
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
			},
			ExpectedError: nil,
		},
		// In this test case, /mnt/baz is a symbolic link that evaluates
		// to /mnt/foo.
		{
			Name: "Mount Point (Symbolic Link)",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/baz",
					},
				},
			},
			GetFile: func(p string) (*model.File, error) {
				switch p {
				case "/mnt/baz":
					return &model.File{Path: "/mnt/foo", Type: model.Directory}, nil
				case "/mnt":
					return &model.File{Path: "/mnt", Type: model.Directory}, nil
				default:
					return nil, utils.NewNotImeplementedError("GetFile()")
				}
			},
			ExpectedOutput: map[string]*model.File{
				"/mnt/baz": {
					Path: "/mnt/foo",
					Type: model.Directory,
				},
				"/mnt": {
					Path: "/mnt",
					Type: model.Directory,
				},
			},
			ExpectedError: nil,
		},
		{
			Name: "Skip + No Mount Point Provided",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {},
				},
			},
			ExpectedOutput: map[string]*model.File{},
			ExpectedError:  nil,
		},
		{
			Name: "Failure to Retrieve File Information",
			Config: &config.Config{
				Devices: map[string]config.Device{
					"/dev/xvdf": {
						MountPoint: "/mnt/foo",
					},
				},
			},
			GetFile: func(p string) (*model.File, error) {
				return nil, fmt.Errorf("🔴 os.stat() failed to retrieve stats for %s", p)
			},
			ExpectedOutput: nil,
			ExpectedError:  fmt.Errorf("🔴 os.stat() failed to retrieve stats for /mnt/foo"),
		},
	}
	for _, subtest := range subtests {
		t.Run(subtest.Name, func(t *testing.T) {
			fs := service.NewMockFileService()
			if subtest.GetFile != nil {
				fs.StubGetFile = subtest.GetFile
			}
			lfb := NewLinuxFileBackend(fs)
			err := lfb.From(subtest.Config)
			utils.CheckError("lfb.From()", t, subtest.ExpectedError, err)
			utils.CheckOutput("lfb.From()", t, subtest.ExpectedOutput, lfb.files)
		})
	}
}
