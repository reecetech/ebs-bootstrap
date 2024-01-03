package service

import (
	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

type MockDeviceService struct {
	StubGetSize         func(name string) (uint64, error)
	StubGetBlockDevices func() ([]string, error)
	StubGetBlockDevice  func(name string) (*model.BlockDevice, error)
	StubMount           func(source string, target string, fs model.FileSystem, options model.MountOptions) error
	StubUmount          func(source string, target string) error
}

func NewMockDeviceService() *MockDeviceService {
	return &MockDeviceService{
		StubGetSize: func(name string) (uint64, error) {
			return 0, utils.NewNotImeplementedError("GetSize()")
		},
		StubGetBlockDevices: func() ([]string, error) {
			return nil, utils.NewNotImeplementedError("GetBlockDevices()")
		},
		StubGetBlockDevice: func(name string) (*model.BlockDevice, error) {
			return nil, utils.NewNotImeplementedError("GetBlockDevice()")
		},
		StubMount: func(source, target string, fs model.FileSystem, options model.MountOptions) error {
			return utils.NewNotImeplementedError("Mount()")
		},
		StubUmount: func(source, target string) error {
			return utils.NewNotImeplementedError("Umount()")
		},
	}
}

func (mds *MockDeviceService) GetSize(name string) (uint64, error) {
	return mds.StubGetSize(name)
}

func (mds *MockDeviceService) GetBlockDevices() ([]string, error) {
	return mds.StubGetBlockDevices()
}

func (mds *MockDeviceService) GetBlockDevice(name string) (*model.BlockDevice, error) {
	return mds.StubGetBlockDevice(name)
}

func (mds *MockDeviceService) Mount(source string, target string, fs model.FileSystem, options model.MountOptions) error {
	return mds.StubMount(source, target, fs, options)
}

func (mds *MockDeviceService) Umount(source string, target string) error {
	return mds.StubUmount(source, target)
}

type MockOwnerService struct {
	StubGetCurrentUser  func() (*model.User, error)
	StubGetCurrentGroup func() (*model.Group, error)
	StubGetUser         func(usr string) (*model.User, error)
	StubGetGroup        func(grp string) (*model.Group, error)
}

func NewMockOwnerService() *MockOwnerService {
	return &MockOwnerService{
		StubGetCurrentUser: func() (*model.User, error) {
			return nil, utils.NewNotImeplementedError("GetCurrentUser()")
		},
		StubGetCurrentGroup: func() (*model.Group, error) {
			return nil, utils.NewNotImeplementedError("GetCurrentGroup()")
		},
		StubGetUser: func(usr string) (*model.User, error) {
			return nil, utils.NewNotImeplementedError("GetUser()")
		},
		StubGetGroup: func(grp string) (*model.Group, error) {
			return nil, utils.NewNotImeplementedError("GetGroup()")
		},
	}
}

func (mos *MockOwnerService) GetCurrentUser() (*model.User, error) {
	return mos.StubGetCurrentUser()
}

func (mos *MockOwnerService) GetCurrentGroup() (*model.Group, error) {
	return mos.StubGetCurrentGroup()
}

func (mos *MockOwnerService) GetUser(usr string) (*model.User, error) {
	return mos.StubGetUser(usr)
}

func (mos *MockOwnerService) GetGroup(grp string) (*model.Group, error) {
	return mos.StubGetGroup(grp)
}

type MockNVMeService struct {
	StubGetBlockDeviceMapping func(device string) (string, error)
}

func NewMockNVMeService() *MockNVMeService {
	return &MockNVMeService{
		StubGetBlockDeviceMapping: func(device string) (string, error) {
			return "", utils.NewNotImeplementedError("GetBlockDeviceMapping()")
		},
	}
}

func (mns *MockNVMeService) GetBlockDeviceMapping(device string) (string, error) {
	return mns.StubGetBlockDeviceMapping(device)
}

// MockFileSystemServiceFactory uses the delegator pattern to inherit any error handling that
// is implemented by FileSystemServiceFactory. This is useful for testing because we can
// stub out the FileSystemService without having to match the error handling logic for FileSystemServiceFactory
// in two places.
type MockFileSystemServiceFactory struct {
	delagator         FileSystemServiceFactory
	fileSystemService FileSystemService
}

func NewMockFileSystemServiceFactory(fssf FileSystemServiceFactory, fss FileSystemService) *MockFileSystemServiceFactory {
	return &MockFileSystemServiceFactory{
		delagator:         fssf,
		fileSystemService: fss,
	}
}

func (mfsf *MockFileSystemServiceFactory) Select(fs model.FileSystem) (FileSystemService, error) {
	_, err := mfsf.delagator.Select(fs)
	if err != nil {
		return nil, err
	}
	return mfsf.fileSystemService, nil
}

type MockFileSystemService struct {
	StubGetSize                 func(name string) (uint64, error)
	StubGetFileSystem           func() model.FileSystem
	StubFormat                  func(name string) error
	StubLabel                   func(name string, label string) error
	StubResize                  func(name string) error
	StubGetMaximumLabelLength   func() int
	StubDoesResizeRequireMount  func() bool
	StubDoesLabelRequireUnmount func() bool
}

func NewMockFileSystemService() *MockFileSystemService {
	return &MockFileSystemService{
		StubGetSize: func(name string) (uint64, error) {
			return 0, utils.NewNotImeplementedError("GetSize()")
		},
		StubGetFileSystem: func() model.FileSystem {
			return model.Unformatted
		},
		StubFormat: func(name string) error {
			return utils.NewNotImeplementedError("Format()")
		},
		StubLabel: func(name string, label string) error {
			return utils.NewNotImeplementedError("Label()")
		},
		StubResize: func(name string) error {
			return utils.NewNotImeplementedError("Resize()")
		},
		StubGetMaximumLabelLength: func() int {
			return 0
		},
		StubDoesResizeRequireMount: func() bool {
			return false
		},
		StubDoesLabelRequireUnmount: func() bool {
			return false
		},
	}
}

func (mfs *MockFileSystemService) GetSize(name string) (uint64, error) {
	return mfs.StubGetSize(name)
}

func (mfs *MockFileSystemService) GetFileSystem() model.FileSystem {
	return mfs.StubGetFileSystem()
}

func (mfs *MockFileSystemService) Format(name string) error {
	return mfs.StubFormat(name)
}

func (mfs *MockFileSystemService) Label(name string, label string) error {
	return mfs.StubLabel(name, label)
}

func (mfs *MockFileSystemService) Resize(name string) error {
	return mfs.StubResize(name)
}

func (mfs *MockFileSystemService) GetMaximumLabelLength() int {
	return mfs.StubGetMaximumLabelLength()
}

func (mfs *MockFileSystemService) DoesResizeRequireMount() bool {
	return mfs.StubDoesResizeRequireMount()
}

func (mfs *MockFileSystemService) DoesLabelRequireUnmount() bool {
	return mfs.StubDoesLabelRequireUnmount()
}

type MockFileService struct {
	StubGetFile           func(file string) (*model.File, error)
	StubCreateDirectory   func(p string) error
	StubChangeOwner       func(p string, uid model.UserId, gid model.GroupId) error
	StubChangePermissions func(p string, perms model.FilePermissions) error
}

func NewMockFileService() *MockFileService {
	return &MockFileService{
		StubGetFile: func(file string) (*model.File, error) {
			return nil, utils.NewNotImeplementedError("GetFile()")
		},
		StubCreateDirectory: func(p string) error {
			return utils.NewNotImeplementedError("CreateDirectory()")
		},
		StubChangeOwner: func(p string, uid model.UserId, gid model.GroupId) error {
			return utils.NewNotImeplementedError("ChangeOwner()")
		},
		StubChangePermissions: func(p string, perms model.FilePermissions) error {
			return utils.NewNotImeplementedError("ChangePermissions()")
		},
	}
}

func (mfs *MockFileService) GetFile(file string) (*model.File, error) {
	return mfs.StubGetFile(file)
}

func (mfs *MockFileService) CreateDirectory(p string) error {
	return mfs.StubCreateDirectory(p)
}

func (mfs *MockFileService) ChangeOwner(p string, uid model.UserId, gid model.GroupId) error {
	return mfs.StubChangeOwner(p, uid, gid)
}

func (mfs *MockFileService) ChangePermissions(p string, perms model.FilePermissions) error {
	return mfs.StubChangePermissions(p, perms)
}
