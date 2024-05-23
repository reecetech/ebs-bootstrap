package backend

import (
	"fmt"
	"os"
	"path"

	"github.com/reecetech/ebs-bootstrap/internal/action"
	"github.com/reecetech/ebs-bootstrap/internal/config"
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/service"
)

type FileBackend interface {
	CreateDirectory(p string) action.Action
	ChangeOwner(p string, uid model.UserId, gid model.GroupId) action.Action
	ChangePermissions(p string, perms model.FilePermissions) action.Action
	GetDirectory(p string) (*model.File, error)
	IsMount(p string) bool
	From(config *config.Config) error
}

type LinuxFileBackend struct {
	files       map[string]*model.File
	fileService service.FileService
}

func NewLinuxFileBackend(fs service.FileService) *LinuxFileBackend {
	return &LinuxFileBackend{
		files:       map[string]*model.File{},
		fileService: fs,
	}
}

func NewMockLinuxFileBackend(files map[string]*model.File) *LinuxFileBackend {
	return &LinuxFileBackend{
		files:       files,
		fileService: nil,
	}
}

func (lfb *LinuxFileBackend) CreateDirectory(p string) action.Action {
	return action.NewCreateDirectoryAction(p, lfb.fileService)
}

func (lfb *LinuxFileBackend) ChangeOwner(p string, uid model.UserId, gid model.GroupId) action.Action {
	return action.NewChangeOwnerAction(p, uid, gid, lfb.fileService)
}

func (lfb *LinuxFileBackend) ChangePermissions(p string, perms model.FilePermissions) action.Action {
	return action.NewChangePermissionsAction(p, perms, lfb.fileService)
}

func (lfb *LinuxFileBackend) GetDirectory(p string) (*model.File, error) {
	f, exists := lfb.files[p]
	if !exists {
		return nil, os.ErrNotExist
	}
	if f.Type != model.Directory {
		return nil, fmt.Errorf("🔴 %s: File is not a directory", p)
	}
	return f, nil
}

// Implementation ported from the Python implementation of os.path.ismount()
// with some minor modifications. For example, we evaluate symbolic links in
// advance, therefore it would be more appropriate to check if the path is
// a directory instead of a symbolic link.
// https://github.com/python/cpython/blob/main/Lib/posixpath.py
func (lfb *LinuxFileBackend) IsMount(p string) bool {
	child, exists := lfb.files[p]
	if !exists {
		return false
	}
	if child.Type != model.Directory {
		return false
	}
	parent, exists := lfb.files[path.Dir(p)]
	if !exists {
		return false
	}
	if child.DeviceId != parent.DeviceId {
		return true
	}
	// If the device IDs are the same, check if the inode numbers of the path
	// and the parent are identical. This condition is true for the root of a filesystem,
	// which is always a mount point.
	return child.InodeNo == parent.InodeNo
}

func (lfb *LinuxFileBackend) From(config *config.Config) error {
	lfb.files = nil
	files := map[string]*model.File{}

	for _, cd := range config.Devices {
		if len(cd.MountPoint) == 0 {
			continue
		}
		// For certain file operations (like lfb.IsMount()) it is essential
		// that we can query the parent directory. Therefore, lets pull the
		// state of the parent directory of the mount point
		mounts := []string{cd.MountPoint, path.Dir(cd.MountPoint)}
		for _, mount := range mounts {
			if _, exists := files[mount]; exists {
				continue
			}
			f, err := lfb.fileService.GetFile(mount)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return err
			}
			files[mount] = f
		}
	}
	lfb.files = files
	return nil
}
