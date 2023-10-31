package state

import (
	"fmt"
	"testing"
	"ebs-bootstrap/internal/utils"
	"ebs-bootstrap/internal/service"
	"ebs-bootstrap/internal/config"
)

type mockDeviceService struct {
	getDeviceInfo func(device string) (*service.DeviceInfo, error)
}

func (ds *mockDeviceService) GetBlockDevices() ([]string, error) {
	return nil, fmt.Errorf("🔴 GetBlockDevices() not implemented")
}

func (ds *mockDeviceService) GetDeviceInfo(device string) (*service.DeviceInfo, error) {
	return ds.getDeviceInfo(device)
}

type mockFileService struct {
	getStats func(file string) (*service.FileInfo, error)
}

func (fs *mockFileService) GetStats(file string) (*service.FileInfo, error) {
	return fs.getStats(file)
}

func (fs *mockFileService) ValidateFile(path string) (error) {
	return fmt.Errorf("🔴 ValidateFile() not implemented")
}

func TestDevice(t *testing.T) {
    subtests := []struct {
		Name		string
        DeviceName      string
		DeviceService	service.DeviceService
		FileService 	service.FileService
		ExpectedErr		error
    }{
        {
			Name:			"Non-Existent Device",
			DeviceName:		"/dev/doesnt-exist",
			DeviceService:	&mockDeviceService{
								getDeviceInfo: func(device string) (*service.DeviceInfo, error) {
									return nil, fmt.Errorf("🔴 /dev/doesnt-exist not a block device")
								},
							},
			FileService:	&mockFileService{
								getStats: func(file string) (*service.FileInfo, error) {
									return nil, fmt.Errorf("🔴 getStats() should not be called")
								},
							},
			ExpectedErr: fmt.Errorf("🔴 /dev/doesnt-exist not a block device"),
        },
    }
    for _, subtest := range subtests {
        t.Run(subtest.Name, func(t *testing.T) {
			_, err := NewDevice(subtest.DeviceName, 
								subtest.DeviceService, 
								subtest.FileService)
			utils.CheckError("NewDevice()", t, subtest.ExpectedErr, err)
        })
    }
}

func TestDeviceDiff(t *testing.T) {
    subtests := []struct {
		Name		string
        DeviceName      string
		DeviceService	service.DeviceService
		FileService 	service.FileService
		Config			*config.Config
		ExpectedErr		error
    }{
        {
			Name:			"No Diff Expected With Mount-Point",
			DeviceName:		"/dev/nvme0n1",
			DeviceService:	&mockDeviceService{
								getDeviceInfo: func(device string) (*service.DeviceInfo, error) {
									return &service.DeviceInfo{
										Name:		"/dev/nvme0n1",
										Label:		"external-vol",
										Fs:			"xfs",
										MountPoint:	"/mnt/app",
									}, nil
								},
							},
			FileService:	&mockFileService{
								getStats: func(file string) (*service.FileInfo, error) {
									return &service.FileInfo{
										Owner:			"100",
										Group:			"100",
										Permissions:	"755",
										Exists:			true,
									}, nil
								},
							},
			Config:			&config.Config{
								Devices: map[string]config.ConfigDevice{
									"/dev/nvme0n1": config.ConfigDevice{
										Fs:			 "xfs",
										MountPoint:	 "/mnt/app",
										Owner: 		 "100",
										Group: 		 "100",
										Label: 		 "external-vol",
										Permissions: "755",
									},
								},
							},
			ExpectedErr: nil,
        },
        {
			Name:			"No Diff Expected Without Mount-Point",
			DeviceName:		"/dev/nvme0n1",
			DeviceService:	&mockDeviceService{
								getDeviceInfo: func(device string) (*service.DeviceInfo, error) {
									return &service.DeviceInfo{
										Name:		"/dev/nvme0n1",
										Label:		"external-vol",
										Fs:			"xfs",
										MountPoint:	"",
									}, nil
								},
							},
			FileService:	&mockFileService{
								getStats: func(file string) (*service.FileInfo, error) {
									return nil, fmt.Errorf("🔴 getStats() should not be called")
								},
							},
			Config:			&config.Config{
								Devices: map[string]config.ConfigDevice{
									"/dev/nvme0n1": config.ConfigDevice{
										Fs:			 "xfs",
										Label: 		 "external-vol",
									},
								},
							},
			ExpectedErr: nil,
        },
        {
			Name:			"Diff Suggesting Fs Change (xfs->ext4)",
			DeviceName:		"/dev/nvme0n1",
			DeviceService:	&mockDeviceService{
								getDeviceInfo: func(device string) (*service.DeviceInfo, error) {
									return &service.DeviceInfo{
										Name:		"/dev/nvme0n1",
										Label:		"external-vol",
										Fs:			"xfs",
										MountPoint:	"/mnt/app",
									}, nil
								},
							},
			FileService:	&mockFileService{
								getStats: func(file string) (*service.FileInfo, error) {
									return &service.FileInfo{
										Owner:			"100",
										Group:			"100",
										Permissions:	"755",
										Exists:			true,
									}, nil
								},
							},
			Config:			&config.Config{
								Devices: map[string]config.ConfigDevice{
									"/dev/nvme0n1": config.ConfigDevice{
										Fs:			 "ext4",
										MountPoint:	 "/mnt/app",
										Owner: 		 "100",
										Group: 		 "100",
										Label: 		 "external-vol",
										Permissions: "755",
									},
								},
							},
			ExpectedErr: 	fmt.Errorf("🔴 File System [/dev/nvme0n1]: Expected=ext4"),
        },
    }
    for _, subtest := range subtests {
        t.Run(subtest.Name, func(t *testing.T) {
			d, err := NewDevice(subtest.DeviceName, 
								subtest.DeviceService, 
								subtest.FileService)
			if err != nil {
				t.Errorf("NewDevice() [error] %s", err)
			}
			err = d.Diff(subtest.Config)
			utils.CheckError("Diff()", t, subtest.ExpectedErr, err)
        })
    }
}
