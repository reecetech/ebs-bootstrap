package service

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/reecetech/ebs-bootstrap/internal/model"
)

const (
	DefaultDirectoryPermissions = os.FileMode(0755)
)

type FileService interface {
	GetFile(file string) (*model.File, error)
	CreateDirectory(path string) error
	ChangeOwner(file string, uid model.UserId, gid model.GroupId) error
	ChangePermissions(file string, perms model.FilePermissions) error
}

type UnixFileService struct{}

func NewUnixFileService() *UnixFileService {
	return &UnixFileService{}
}

// We simplify our model of a file system by evaluating symbolic links in advance.
// This means that any symbolic links will be resolved to a *model.File that reflects
// its target.
// This behaviour is useful because if a device is mounted to a symbolic link
// (or any nested directory) using `mount`, the `lsblk` tool will report the
// resolved location of the symbolic link.
func (ufs *UnixFileService) GetFile(file string) (*model.File, error) {
	info, err := os.Stat(file)
	if err != nil {
		return nil, err
	}

	var ft model.FileType
	switch mode := info.Mode(); {
	case mode.IsRegular():
		ft = model.RegularFile
	case mode.IsDir():
		ft = model.Directory
	default:
		ft = model.Special
	}

	file, err = filepath.EvalSymlinks(file)
	if err != nil {
		return nil, err
	}

	stat, ok := info.Sys().(*syscall.Stat_t)
	if ok {
		return &model.File{
			Path:        file,
			DeviceId:    stat.Dev,
			InodeNo:     stat.Ino,
			UserId:      model.UserId(stat.Uid),
			GroupId:     model.GroupId(stat.Gid),
			Permissions: model.FilePermissions(info.Mode().Perm()),
			Type:        ft,
		}, nil
	}
	return nil, fmt.Errorf("🔴 %s: Failed to get os.stat() information", file)
}

func (ufs *UnixFileService) CreateDirectory(path string) error {
	return os.MkdirAll(path, DefaultDirectoryPermissions)
}

func (ufs *UnixFileService) ChangeOwner(file string, uid model.UserId, gid model.GroupId) error {
	return os.Chown(file, int(uid), int(gid))
}

func (ufs *UnixFileService) ChangePermissions(file string, perms model.FilePermissions) error {
	return os.Chmod(file, perms.Perm())
}
