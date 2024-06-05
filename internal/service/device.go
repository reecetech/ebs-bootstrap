package service

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/reecetech/ebs-bootstrap/internal/model"
	"github.com/reecetech/ebs-bootstrap/internal/utils"
)

var deviceNameRegex = regexp.MustCompile(`^NAME="(.*)"`)
var devicePropertiesRegex = regexp.MustCompile(`^LABEL="(.*)" FSTYPE="(.*)" MOUNTPOINT="(.*)"`)

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
	output, err := r.Command("--nodeps", "-o", "NAME", "-P")
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	d := make([]string, len(lines))

	for i, line := range lines {
		matches := deviceNameRegex.FindStringSubmatch(line)
		if len(matches) != 2 {
			return nil, fmt.Errorf("🔴 Failed to decode lsblk response")
		}
		d[i] = "/dev/" + matches[1]
	}
	return d, nil
}

func (du *LinuxDeviceService) GetBlockDevice(name string) (*model.BlockDevice, error) {
	r := du.runnerFactory.Select(utils.Lsblk)
	output, err := r.Command("--nodeps", "-o", "LABEL,FSTYPE,MOUNTPOINT", "-P", name)
	if err != nil {
		return nil, err
	}

	matches := devicePropertiesRegex.FindStringSubmatch(output)
	if len(matches) != 4 {
		return nil, fmt.Errorf("🔴 Failed to decode lsblk response")
	}

	fs, err := model.ParseFileSystem(matches[2])
	if err != nil {
		return nil, fmt.Errorf("🔴 %s: %s", name, err)
	}
	return &model.BlockDevice{
		Name:       name,
		Label:      matches[1],
		FileSystem: fs,
		MountPoint: matches[3],
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
