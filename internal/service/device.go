package service

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

type DeviceService interface {
	GetSize(name string) (uint64, error) // bytes
	GetBlockDevices() ([]string, error)
	GetBlockDevice(name string) (*model.BlockDevice, error)
	Mount(source string, target string, fs model.FileSystem, options model.MountOptions) error
	Umount(source string, target string) error
}

type LinuxDeviceService struct {
	runnerFactory utils.RunnerFactory
}

type LsblkBlockDeviceResponse struct {
	BlockDevices []struct {
		Name       *string `json:"name"`
		Label      *string `json:"label"`
		FsType     *string `json:"fstype"`
		MountPoint *string `json:"mountpoint"`
	} `json:"blockdevices"`
}

func NewLinuxDeviceService(rc utils.RunnerFactory) *LinuxDeviceService {
	return &LinuxDeviceService{
		runnerFactory: rc,
	}
}

func (du *LinuxDeviceService) GetSize(name string) (uint64, error) {
	r := du.runnerFactory.Select(utils.BlockDev)
	output, err := r.Command("--getsize64", name)
	if err != nil {
		return 0, err
	}
	b, err := strconv.ParseUint(output, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("🔴 Failed to cast block device size to unsigned 64-bit integer")
	}
	return b, nil
}

func (du *LinuxDeviceService) GetBlockDevices() ([]string, error) {
	r := du.runnerFactory.Select(utils.Lsblk)
	output, err := r.Command("--nodeps", "-o", "NAME", "-J")
	if err != nil {
		return nil, err
	}
	lbd := &LsblkBlockDeviceResponse{}
	err = json.Unmarshal([]byte(output), lbd)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to decode lsblk response: %v", err)
	}
	d := make([]string, len(lbd.BlockDevices))
	for i := range d {
		d[i] = "/dev/" + utils.Safe(lbd.BlockDevices[i].Name)
	}
	return d, nil
}

func (du *LinuxDeviceService) GetBlockDevice(name string) (*model.BlockDevice, error) {
	r := du.runnerFactory.Select(utils.Lsblk)
	output, err := r.Command("--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-J", name)
	if err != nil {
		return nil, err
	}
	lbd := &LsblkBlockDeviceResponse{}
	err = json.Unmarshal([]byte(output), lbd)
	if err != nil {
		return nil, fmt.Errorf("🔴 Failed to decode lsblk response: %v", err)
	}
	if len(lbd.BlockDevices) != 1 {
		return nil, fmt.Errorf("🔴 %s: An unexpected number of block devices were returned: Expected=1 Actual=%d", name, len(lbd.BlockDevices))
	}
	fst := utils.Safe(lbd.BlockDevices[0].FsType)
	fs, err := model.ParseFileSystem(fst)
	if err != nil {
		return nil, fmt.Errorf("🔴 %s: %s", name, err)
	}
	return &model.BlockDevice{
		Name:       name,
		Label:      utils.Safe(lbd.BlockDevices[0].Label),
		FileSystem: fs,
		MountPoint: utils.Safe(lbd.BlockDevices[0].MountPoint),
	}, nil
}

func (du *LinuxDeviceService) Mount(source string, target string, fs model.FileSystem, options model.MountOptions) error {
	r := du.runnerFactory.Select(utils.Mount)
	_, err := r.Command(source, "-t", string(fs), "-o", string(options), target)
	return err
}

func (du *LinuxDeviceService) Umount(source string, target string) error {
	r := du.runnerFactory.Select(utils.Umount)
	_, err := r.Command(target)
	return err
}
