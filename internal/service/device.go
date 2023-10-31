package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"ebs-bootstrap/internal/utils"
)

// Device Service Interface [START]

type DeviceInfo struct {
	Name		string
	Label		string
	Fs			string
	MountPoint	string
}

type DeviceService interface {
	GetBlockDevices() ([]string, error)
	GetDeviceInfo(device string) (*DeviceInfo, error)
}

// Device Service Interface [END]

type LinuxDeviceService struct {
	Runner utils.Runner
}

type LsblkBlockDeviceResponse struct {
	BlockDevices	[]LsblkBlockDevice `json:"blockdevices"`
}

type LsblkBlockDevice struct {
	Name		string	`json:"name"`
	Label		string	`json:"label"`
	FsType		string	`json:"fstype"`
	MountPoint	string	`json:"mountpoint"`
}

func (du *LinuxDeviceService) GetBlockDevices() ([]string, error) {
	output, err := du.Runner.Command("lsblk", "--nodeps", "-o", "NAME,LABEL,FSTYPE,MOUNTPOINT", "-J")
	if err != nil {
		return nil, err
	}
	lbd := &LsblkBlockDeviceResponse{}
	err = json.Unmarshal([]byte(output), lbd)
	if err != nil {
		return nil, err
	}
	d := make([]string,len(lbd.BlockDevices))
	for i, _ := range d {
		d[i] = "/dev/" + lbd.BlockDevices[i].Name
	}
	return d, nil
}

func (du *LinuxDeviceService) GetDeviceInfo(device string) (*DeviceInfo, error) {
	output, err := du.Runner.Command("lsblk", "--nodeps", "-o", "NAME,LABEL,FSTYPE,MOUNTPOINT", "-J", device)
	if err != nil {
		return nil, err
	}
	bd := &LsblkBlockDeviceResponse{}
	err = json.Unmarshal([]byte(output), bd)
	if err != nil {
		return nil, err
	}
	if len(bd.BlockDevices) != 1 {
		return nil, fmt.Errorf("🔴 [%s] An unexpected number of block devices were returned: Expected=1 Actual=%d", device, len(bd.BlockDevices))
	}
	return &DeviceInfo{
		Name: "/dev/" + bd.BlockDevices[0].Name,
		Label: bd.BlockDevices[0].Label,
		Fs: bd.BlockDevices[0].FsType,
		MountPoint: bd.BlockDevices[0].MountPoint,
	}, nil
}

// Device Translator Service Interface [START]

type DeviceTranslator struct {
	Table map[string]string
}

type DeviceTranslatorService interface {
	GetTranslator() *DeviceTranslator
}

type EbsDeviceTranslator struct {
	DeviceService DeviceService
	NVMeService NVMeService
}

// Device Translator Service Interface [END]

func (edt *EbsDeviceTranslator) GetTranslator() (*DeviceTranslator, error) {
	dt := &DeviceTranslator{}
	dt.Table = make(map[string]string)
	devices, err := edt.DeviceService.GetBlockDevices()
	if err != nil {
		return nil, err
	}
	for _, device := range(devices) {
		alias := device
		if strings.HasPrefix(device, "/dev/nvme") {
			alias, err = edt.NVMeService.GetBlockDeviceMapping(device)
			if err != nil {
				return nil, err
			}
		}
		dt.Table[alias] = device
		dt.Table[device] = alias
	}
	return dt, nil
}
